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
	tag              string
	content          fyne.CanvasObject
	child            fyne.CanvasObject
	lastClick        time.Time
	visible          bool
	disabled         bool
	dragScroll       float32
	dragPosition     fyne.Position
	dragOffset       fyne.Position
	dragMoveEnd      fyne.Position
	dragMoveStart    fyne.Position
	dragCursorOffset fyne.Position
	dragging         bool
	background       *canvas.Rectangle
	refTitle         *canvas.Text
	refContent       *canvas.Text
	refSubtitle      *canvas.Text
	refData          *JT.PanelDataType
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
			if SyncingPanels || !Draggable {
				return
			}
			dynpk, _ := str.Get()
			if onEdit != nil {
				go onEdit(dynpk, uuid)
			}
		},
		func() {
			if SyncingPanels || !Draggable {
				return
			}
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
		child:       action,
		visible:     false,
		disabled:    false,
		background:  background,
		refTitle:    title,
		refContent:  content,
		refSubtitle: subtitle,
		refData:     pdt,
	}

	panel.ExtendBaseWidget(panel)
	action.Hide()

	str.AddListener(binding.NewDataListener(func() {
		if !panel.refData.DidChange() && panel.refData.Status == 1 {
			return
		}

		switch panel.refData.IsValueIncrease() {
		case 1:
			panel.background.FillColor = JC.GreenColor
			panel.background.Refresh()
		case -1:
			panel.background.FillColor = JC.RedColor
			panel.background.Refresh()
		}

		panel.updateContent()
		JA.StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	panel.updateContent()
	return panel
}

func (h *PanelDisplay) updateContent() {
	if h.refData.UsePanelKey().GetValueFloat() != -1 {
		h.refData.Status = 1
	}

	switch h.refData.Status {
	case -1:
		h.refTitle.Text = "Fetching Rates..."
		h.refSubtitle.Hide()
		h.refContent.Hide()
		h.DisableClick()
		h.background.FillColor = JC.PanelBG

	case 0:
		h.refTitle.Text = "Loading..."
		h.refSubtitle.Hide()
		h.refContent.Hide()
		h.DisableClick()
		h.background.FillColor = JC.PanelBG

	default:
		if !JT.BP.ValidatePanel(h.refData.Get()) {
			h.refTitle.Text = "Invalid Panel"
			h.refSubtitle.Hide()
			h.refContent.Hide()
			h.background.FillColor = JC.PanelBG
			return
		}

		h.refTitle.Text = h.refData.FormatTitle()
		h.refSubtitle.Text = h.refData.FormatSubtitle()
		h.refContent.Text = h.refData.FormatContent()

		h.refSubtitle.Show()
		h.refContent.Show()
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

	scrollY := JM.AppMainPanelScrollWindow.Offset.Y

	if !h.dragging {
		h.dragging = true

		// Offset between cursor and widget's top-left corner
		h.dragCursorOffset = ev.Position.Subtract(h.Position())

		// Initial move — use live scroll offset
		h.dragScroll = JM.AppMainPanelScrollWindow.Offset.Y
		initialPos := fyne.NewPos(
			ev.Position.X-h.dragCursorOffset.X,
			ev.Position.Y-h.dragCursorOffset.Y,
		)

		DragPlaceholder.Move(initialPos)
		JC.Grid.Add(DragPlaceholder)

		// Handling scroll event when still in drag mode
		go func() {
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			for h.dragging {
				<-ticker.C
				currentScroll := JM.AppMainPanelScrollWindow.Offset.Y
				if currentScroll != scrollY {
					newPos := fyne.NewPos(
						h.dragPosition.X-h.dragCursorOffset.X,
						h.dragPosition.Y-h.dragCursorOffset.Y+(currentScroll-h.dragScroll),
					)

					scrollY = currentScroll

					fyne.Do(func() {
						DragPlaceholder.Move(newPos)
					})
				}
			}
		}()
	}

	newPos := fyne.NewPos(
		ev.Position.X-h.dragCursorOffset.X,
		ev.Position.Y-h.dragCursorOffset.Y,
	)

	if scrollY != h.dragScroll {
		newPos.Y += scrollY - h.dragScroll
	}

	DragPlaceholder.Move(newPos)

	// Store the final position for snapping and reordering later!
	h.dragPosition = ev.Position
}

func (h *PanelDisplay) DragEnd() {
	// Call this early to cancel go routine
	h.dragging = false
	JC.Grid.Remove(DragPlaceholder)

	h.dragOffset = h.Position().Add(h.dragPosition)
	h.dragOffset.Y -= h.dragScroll - JM.AppMainPanelScrollWindow.Offset.Y

	h.snapToNearest()
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

	JC.Logln(fmt.Sprintf("Dragging item - Position: (%.2f, %.2f)", dragPos.X, dragPos.Y))

	for i, panel := range panels {
		panelPos := panel.Position()
		panelSize := panel.Size()

		left := panelPos.X
		right := panelPos.X + panelSize.Width
		top := panelPos.Y
		bottom := panelPos.Y + panelSize.Height

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

	JA.FadeInBackground(h.background, 300*time.Millisecond)

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

	JT.BP.Set(nd)

	return true
}
