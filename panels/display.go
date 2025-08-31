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

var activeAction *PanelDisplay = nil
var activeDragging *PanelDisplay = nil

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
	spacer := float32(-3)

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

	totalHeight += spacer * float32(len(centerItems)-1)

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	for _, obj := range centerItems {
		objSize := obj.MinSize()
		obj.Move(fyne.NewPos((size.Width-objSize.Width)/2, currentY))
		obj.Resize(objSize)
		currentY += objSize.Height + spacer
	}

	actionSize := action.MinSize()
	action.Move(fyne.NewPos(size.Width-actionSize.Width, 0))
	action.Resize(actionSize)

	if JM.AppStatusManager.IsDraggable() {
		action.Hide()
		if activeAction != nil {
			activeAction.visible = false
			activeAction = nil
		}
	}
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
	tag          string
	content      fyne.CanvasObject
	child        fyne.CanvasObject
	container    *PanelGridContainer
	lastClick    time.Time
	visible      bool
	disabled     bool
	dragScroll   float32
	dragPosition fyne.Position
	dragOffset   fyne.Position
	dragging     bool
	background   *canvas.Rectangle
	refTitle     *canvas.Text
	refContent   *canvas.Text
	refSubtitle  *canvas.Text
	fps          time.Duration
}

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
				go JA.FadeOutBackground(background, 300*time.Millisecond, func() {
					onDelete(uuid)
				})
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
	}

	panel.fps = 8 * time.Millisecond
	if JC.IsMobile {
		panel.fps = 32 * time.Millisecond
	}

	panel.ExtendBaseWidget(panel)
	action.Hide()

	str.AddListener(binding.NewDataListener(func() {

		pkt := JT.BP.GetData(panel.GetTag())
		if pkt == nil {
			return
		}

		if !pkt.DidChange() && pkt.Status == 1 {
			return
		}

		switch pkt.IsValueIncrease() {
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

	JA.FadeInBackground(background, 300*time.Millisecond, nil)

	return panel
}

func (h *PanelDisplay) updateContent() {

	pwidth := h.Size().Width
	if pwidth != 0 && pwidth < JC.PanelWidth {
		h.refTitle.TextSize = JC.PanelTitleSizeSmall
		h.refSubtitle.TextSize = JC.PanelSubTitleSizeSmall
		h.refContent.TextSize = JC.PanelContentSizeSmall
	}

	pkt := JT.BP.GetData(h.GetTag())
	if pkt == nil {
		return
	}

	if pkt.UsePanelKey().GetValueFloat() != -1 {
		pkt.Status = 1
	}

	switch pkt.Status {
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
		if !JT.BP.ValidatePanel(pkt.Get()) {
			h.refTitle.Text = "Invalid Panel"
			h.refSubtitle.Hide()
			h.refContent.Hide()
			h.background.FillColor = JC.PanelBG
			return
		}

		h.refTitle.Text = JC.TruncateText(pkt.FormatTitle(), pwidth-20, h.refTitle.TextSize)
		h.refSubtitle.Text = JC.TruncateText(pkt.FormatSubtitle(), pwidth-20, h.refSubtitle.TextSize)
		h.refContent.Text = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.refContent.TextSize)

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
	if JM.AppStatusManager.IsDraggable() {
		h.HideTarget()
		return
	}

	if h.visible && !h.child.Visible() {
		h.ShowTarget()
		return
	}

	if activeDragging != nil {
		return
	}

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
	if JM.AppStatusManager.IsDraggable() {
		return desktop.PointerCursor
	}

	return desktop.DefaultCursor
}

func (h *PanelDisplay) DisableClick() {
	h.disabled = true
}

func (h *PanelDisplay) EnableClick() {
	h.disabled = false
}

func (h *PanelDisplay) PanelDrag(ev *fyne.DragEvent) {
	if activeDragging != nil && activeDragging != h {
		activeDragging.DragEnd()
	}

	h.dragPosition = ev.Position

	if activeDragging == nil {
		activeDragging = h
		h.dragging = true
		h.dragScroll = JM.AppLayoutManager.OffsetY()

		p := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		JM.DragPlaceholder.Move(p)
		canvas.Refresh(JM.DragPlaceholder)

		go func() {
			ticker := time.NewTicker(h.fps)
			defer ticker.Stop()

			rect, ok := JM.DragPlaceholder.(*canvas.Rectangle)

			if !ok {
				return
			}

			shown := false
			posX := rect.Position().X
			posY := rect.Position().Y
			placeholderSize := rect.Size()
			edgeThreshold := placeholderSize.Height / 2
			scrollStep := float32(10)

			for h.dragging {
				<-ticker.C

				targetX := p.X + h.dragPosition.X - (placeholderSize.Width / 2)
				targetY := p.Y + h.dragPosition.Y - (placeholderSize.Height / 2)

				edgeTopY := JM.AppLayoutManager.ContentTopY - edgeThreshold
				edgeBottomY := JM.AppLayoutManager.ContentBottomY - edgeThreshold

				// Show placeholder
				if !shown && (targetX == posX || targetY == posY) {
					fyne.Do(func() {
						rect.FillColor = JC.PanelPlaceholderBG
						canvas.Refresh(rect)
						shown = true
					})
				}

				// Move placeholder
				if posX != targetX || posY != targetY {
					posX = targetX
					posY = targetY
					fyne.Do(func() {
						rect.Move(fyne.NewPos(posX, posY))
						canvas.Refresh(rect)
					})
				}

				// Scroll when placeholder is half out of viewport
				if posY < edgeTopY {
					JM.AppLayoutManager.ScrollBy(-scrollStep)
				} else if posY > edgeBottomY {
					JM.AppLayoutManager.ScrollBy(scrollStep)
				}
			}
		}()
	}
}

func (h *PanelDisplay) Dragged(ev *fyne.DragEvent) {
	if JM.AppStatusManager.IsDraggable() {
		h.PanelDrag(ev)
	} else {
		h.container.Dragged(ev)
	}
}

func (h *PanelDisplay) DragEnd() {
	if !JM.AppStatusManager.IsDraggable() {
		h.dragging = false
		h.container.DragEnd()
		return
	}

	// Call this early to cancel go routine
	activeDragging = nil
	h.dragging = false

	if rect, ok := JM.DragPlaceholder.(*canvas.Rectangle); ok {
		rect.FillColor = JC.Transparent
		rect.Move(fyne.NewPos(0, -JC.PanelHeight))
		canvas.Refresh(rect)
	}

	h.dragOffset = h.Position().Add(h.dragPosition)
	h.dragOffset.Y -= h.dragScroll - JM.AppLayoutManager.OffsetY()

	h.snapToNearest()
}

func (h *PanelDisplay) snapToNearest() {

	// Convert target position to grid index
	targetIndex := h.findDropTargetIndex()
	if targetIndex == -1 {
		return
	}

	Grid.Objects = h.reorder(targetIndex)
	Grid.Refresh()

	go func() {
		JC.MainDebouncer.Call("panel_drag", 1000*time.Millisecond, func() {
			if h.syncPanelData() {
				if JT.SavePanels() {
					JC.Notify("Panels have been reordered and updated.")
				}
			}
		})
	}()
}

func (h *PanelDisplay) findDropTargetIndex() int {

	JC.Logln(fmt.Sprintf("Dragging item - Position: (%.2f, %.2f)", h.dragOffset.X, h.dragOffset.Y))

	for i, zone := range DragDropZones {

		JC.Logln(fmt.Sprintf(
			"Checking panel %d — Bounds: [X: %.2f–%.2f, Y: %.2f–%.2f]",
			i, zone.left, zone.right, zone.top, zone.bottom,
		))

		if h.dragOffset.X >= zone.left &&
			h.dragOffset.X <= zone.right &&
			h.dragOffset.Y >= zone.top &&
			h.dragOffset.Y <= zone.bottom {

			if zone.panel == h {
				JC.Logln(fmt.Sprintf("Refusing to drop panel to the old position %d", i))
				return -1
			}

			JC.Logln(fmt.Sprintf("Dropped inside panel %d", i))

			return i
		}
	}

	JC.Logln("Refuse to drop panel to invalid drop position")

	return -1
}

func (h *PanelDisplay) reorder(targetIndex int) []fyne.CanvasObject {
	panels := Grid.Objects
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

	JA.FadeInBackground(h.background, 300*time.Millisecond, nil)

	return result
}

func (h *PanelDisplay) syncPanelData() bool {
	nd := []*JT.PanelDataType{}

	for _, obj := range Grid.Objects {
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
