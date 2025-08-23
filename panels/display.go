package panels

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JM "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var Draggable = true
var SyncingPanels = false

type PanelLayout struct{}

func (p *PanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 5 {
		return
	}
	bg := objects[0]
	title := objects[1]
	content := objects[2]
	subtitle := objects[3]
	action := objects[4]

	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{title, content, subtitle} {
		if obj.Visible() && obj.MinSize().Height > 0 {
			centerItems = append(centerItems, obj)
		}
	}

	var totalHeight float32
	for _, obj := range centerItems {
		totalHeight += obj.MinSize().Height
	}

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	for _, obj := range centerItems {
		objSize := obj.MinSize()
		obj.Move(fyne.NewPos((size.Width-objSize.Width)/2, currentY))
		obj.Resize(objSize)
		currentY += objSize.Height
	}

	actionSize := action.MinSize()
	action.Move(fyne.NewPos(size.Width-actionSize.Width, 0))
	action.Resize(actionSize)
}

func (p *PanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	for _, obj := range objects[1:4] {
		if obj.Visible() && obj.MinSize().Height > 0 {
			size := obj.MinSize()
			if size.Width > width {
				width = size.Width
			}
			height += size.Height
		}
	}

	return fyne.NewSize(width, height)
}

type PanelDisplay struct {
	widget.BaseWidget
	tag         string
	content     fyne.CanvasObject
	child       fyne.CanvasObject
	lastClick   time.Time
	visible     bool
	disabled    bool
	firstDrag   fyne.Position
	lastDrag    fyne.Position
	startScroll float32
	endScroll   float32
	dragOffset  fyne.Position
	dragging    bool
	background  *canvas.Rectangle
}

var activeAction *PanelDisplay = nil

func NewPanelDisplay(
	pdt *JT.PanelDataType,
	onEdit func(pk string, uuid string),
	onDelete func(uuid string),
) *PanelDisplay {
	uuid := JC.CreateUUID()
	pdt.ID = uuid

	title := canvas.NewText("", JC.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = JC.PanelTitleSize

	subtitle := canvas.NewText("", JC.TextColor)
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = JC.PanelSubTitleSize

	content := canvas.NewText("", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = JC.PanelContentSize

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	str := pdt.GetData()

	action := NewPanelActionBar(
		func() {
			dynpk, _ := str.Get()
			if onEdit != nil {
				go onEdit(dynpk, uuid)
			}
		},
		func() {
			if onDelete != nil {
				go onDelete(uuid)
			}
		},
	)

	panel := &PanelDisplay{
		tag: uuid,
		content: container.New(&PanelLayout{},
			background,
			title,
			content,
			subtitle,
			action,
		),
		child:      action,
		visible:    false,
		disabled:   false,
		background: background,
	}

	panel.ExtendBaseWidget(panel)
	action.Hide()

	str.AddListener(binding.NewDataListener(func() {
		if !pdt.DidChange() && pdt.Status == 1 {
			return
		}

		switch pdt.IsValueIncrease() {
		case 1:
			panel.background.FillColor = JC.GreenColor
			panel.background.Refresh()
		case -1:
			panel.background.FillColor = JC.RedColor
			panel.background.Refresh()
		}

		panel.updateContent(pdt, title, subtitle, content)
		JA.StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	panel.updateContent(pdt, title, subtitle, content)
	return panel
}

func (h *PanelDisplay) updateContent(pdt *JT.PanelDataType, title, subtitle, content *canvas.Text) {
	if pdt.UsePanelKey().GetValueFloat() != -1 {
		pdt.Status = 1
	}

	switch pdt.Status {
	case -1:
		title.Text = "Fetching Rates..."
		subtitle.Hide()
		content.Hide()
		h.DisableClick()
		h.background.FillColor = JC.PanelBG

	case 0:
		title.Text = "Loading..."
		subtitle.Hide()
		content.Hide()
		h.DisableClick()
		h.background.FillColor = JC.PanelBG

	default:
		if !JT.BP.ValidatePanel(pdt.Get()) {
			title.Text = "Invalid Panel"
			subtitle.Hide()
			content.Hide()
			h.background.FillColor = JC.PanelBG
			return
		}

		title.Text = pdt.FormatTitle()
		subtitle.Text = pdt.FormatSubtitle()
		content.Text = pdt.FormatContent()

		subtitle.Show()
		content.Show()
		h.EnableClick()
	}
}

func (h *PanelDisplay) GetTag() string {
	return h.tag
}

func (h *PanelDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *PanelDisplay) Tapped(event *fyne.PointEvent) {
	if h.disabled {
		return
	}

	if h.visible {
		h.HideTarget()
	} else {
		h.ShowTarget()
	}
}

func (h *PanelDisplay) ShowTarget() {
	h.child.Show()
	h.visible = true
	h.Refresh()

	if activeAction != nil {
		activeAction.HideTarget()
	}

	activeAction = h
}

func (h *PanelDisplay) HideTarget() {
	h.child.Hide()
	h.visible = false
	h.Refresh()
	activeAction = nil
}

func (h *PanelDisplay) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (h *PanelDisplay) DisableClick() {
	h.disabled = true
}

func (h *PanelDisplay) EnableClick() {
	h.disabled = false
}

func (h *PanelDisplay) Dragged(ev *fyne.DragEvent) {
	if !Draggable {
		return
	}

	if !h.dragging {
		h.firstDrag = h.Position().Add(ev.Position)
		h.startScroll = JM.AppMainPanelScrollWindow.Offset.Y
		h.dragging = true
	}

	h.lastDrag = ev.Position
}

func (h *PanelDisplay) DragEnd() {
	h.endScroll = JM.AppMainPanelScrollWindow.Offset.Y
	h.dragOffset = h.firstDrag.Add(h.lastDrag)
	h.dragOffset.Y -= h.startScroll - h.endScroll

	h.snapToNearest()

	h.dragging = false
	h.dragOffset = fyne.NewPos(0, 0)
	h.startScroll = 0
	h.endScroll = 0
	h.firstDrag = fyne.NewPos(0, 0)
	h.lastDrag = fyne.NewPos(0, 0)

}

func (h *PanelDisplay) snapToNearest() {

	// Convert target position to grid index
	targetIndex := h.findDropTargetIndex()
	if targetIndex != -1 {
		Draggable = false
		JC.Grid.Objects = h.reorder(targetIndex)
		JC.Grid.Refresh()

		if !SyncingPanels {
			SyncingPanels = true

			go func() {
				time.Sleep(1000 * time.Millisecond)

				if h.syncPanelData() {
					if JT.SavePanels() {
						JC.Notify("Panels updated")
						SyncingPanels = false
					}
				}
			}()
		}

		go func() {
			time.Sleep(1 * time.Millisecond)
			Draggable = true
		}()
	}
}

func (h *PanelDisplay) findDropTargetIndex() int {
	panels := JC.Grid.Objects
	dragPos := h.dragOffset

	JC.Logln(fmt.Sprintf("Dragging item - Position: (%.2f, %.2f) - Offset: (%.2f, %.2f)", dragPos.X, dragPos.Y, h.dragOffset.X, h.dragOffset.Y))

	layout, ok := JC.Grid.Layout.(*PanelGridLayout)
	if !ok {
		return -1
	}

	for i, panel := range panels {
		panelPos := panel.Position()
		panelSize := panel.Size()

		left := panelPos.X - layout.InnerPadding[1]
		right := panelPos.X + panelSize.Width + layout.InnerPadding[3]
		top := panelPos.Y - layout.InnerPadding[0]
		bottom := panelPos.Y + panelSize.Height + layout.InnerPadding[2]

		JC.Logln(fmt.Sprintf(
			"Checking panel %d — Bounds: [X: %.2f–%.2f, Y: %.2f–%.2f]",
			i, left, right, top, bottom,
		))

		if dragPos.X >= left && dragPos.X <= right &&
			dragPos.Y >= top && dragPos.Y <= bottom {

			JC.Logln(fmt.Sprintf("Dropped inside panel %d", i))

			return i
		}
	}

	return -1
}

func (h *PanelDisplay) reorder(targetIndex int) []fyne.CanvasObject {
	panels := JC.Grid.Objects
	var result []fyne.CanvasObject
	for _, obj := range panels {
		if obj != h {
			result = append(result, obj)
		}
	}

	if targetIndex >= len(result) {
		result = append(result, h)
	} else {
		result = append(result[:targetIndex], append([]fyne.CanvasObject{h}, result[targetIndex:]...)...)
	}

	JA.FadeInBackground(h.background, 100*time.Millisecond)

	return result
}

func (h *PanelDisplay) syncPanelData() bool {
	nd := []*JT.PanelDataType{}

	for _, obj := range JC.Grid.Objects {
		if panel, ok := obj.(*PanelDisplay); ok {
			uuid := panel.GetTag()
			pdt := JT.BP.GetData(uuid)
			if pdt != nil {
				nd = append(nd, pdt)
			}
		}
	}

	JT.BP.Inject(nd)

	return true
}
