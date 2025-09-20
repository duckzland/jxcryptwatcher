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

var ActiveAction *PanelDisplay = nil
var activeDragging *PanelDisplay = nil

type PanelDisplay struct {
	widget.BaseWidget
	tag           string
	content       fyne.CanvasObject
	child         fyne.CanvasObject
	container     *panelContainer
	visible       bool
	disabled      bool
	dragScroll    float32
	dragPosition  fyne.Position
	dragOffset    fyne.Position
	dragging      bool
	background    *canvas.Rectangle
	refTitle      *canvas.Text
	refContent    *canvas.Text
	refSubtitle   *canvas.Text
	refBottomText *canvas.Text
	fps           time.Duration
	status        int
}

func NewPanelDisplay(
	pdt *JT.PanelDataType,
	onEdit func(pk string, uuid string),
	onDelete func(uuid string),
) *PanelDisplay {

	uuid := JC.CreateUUID()
	pdt.SetID(uuid)
	str := pdt.GetData()

	pl := &panelLayout{
		title:      canvas.NewText("", JC.TextColor),
		subtitle:   canvas.NewText("", JC.TextColor),
		content:    canvas.NewText("", JC.TextColor),
		background: canvas.NewRectangle(JC.PanelBG),
		bottomText: canvas.NewText("", JC.TextColor),
	}

	pl.title.Alignment = fyne.TextAlignCenter
	pl.title.TextStyle = fyne.TextStyle{Bold: true}
	pl.title.TextSize = JC.PanelTitleSize

	pl.subtitle.Alignment = fyne.TextAlignCenter
	pl.subtitle.TextSize = JC.PanelSubTitleSize

	pl.content.Alignment = fyne.TextAlignCenter
	pl.content.TextStyle = fyne.TextStyle{Bold: true}
	pl.content.TextSize = JC.PanelContentSize

	pl.background.SetMinSize(fyne.NewSize(100, 100))
	pl.background.CornerRadius = JC.PanelBorderRadius

	pl.bottomText.Alignment = fyne.TextAlignCenter
	pl.bottomText.TextSize = JC.PanelBottomTextSize

	pl.action = NewPanelAction(
		func() {
			dynpk, _ := str.Get()
			if onEdit != nil {
				onEdit(dynpk, uuid)
			}
		},
		func() {
			if onDelete != nil {
				JA.FadeOutBackground(pl.background, 300*time.Millisecond, func() {
					onDelete(uuid)
				})
			}
		},
	)

	panel := &PanelDisplay{
		tag: uuid,
		content: container.New(
			pl,
			pl.background,
			pl.title,
			pl.content,
			pl.subtitle,
			pl.bottomText,
			pl.action,
		),
		child:         pl.action,
		visible:       false,
		disabled:      false,
		background:    pl.background,
		refTitle:      pl.title,
		refContent:    pl.content,
		refSubtitle:   pl.subtitle,
		refBottomText: pl.bottomText,
	}

	panel.fps = 3 * time.Millisecond
	if JC.IsMobile {
		panel.fps = 6 * time.Millisecond
	}

	panel.ExtendBaseWidget(panel)
	pl.action.Hide()

	panel.status = pdt.GetStatus()

	str.AddListener(binding.NewDataListener(func() {
		pkt := JT.BP.GetDataByID(panel.GetTag())
		if pkt == nil {
			return
		}

		if pkt.IsStatus(JC.STATE_LOADED) {
			if pkt.DidChange() {
				switch pkt.IsValueIncrease() {
				case JC.VALUE_INCREASE:
					panel.background.FillColor = JC.GreenColor
					panel.background.Refresh()
				case JC.VALUE_DECREASE:
					panel.background.FillColor = JC.RedColor
					panel.background.Refresh()
				}
			} else if pkt.IsOnInitialValue() {
				panel.background.FillColor = JC.PanelBG
			} else if panel.status == JC.STATE_ERROR {
				panel.background.FillColor = JC.PanelBG
			}
		}

		panel.UpdateContent()
		JA.StartFlashingText(panel.refContent, 50*time.Millisecond, JC.TextColor, 1)
	}))

	panel.UpdateContent()
	JA.FadeInBackground(panel.background, 100*time.Millisecond, nil)

	return panel
}

func (h *PanelDisplay) UpdateContent() {

	pwidth := h.Size().Width
	if pwidth != 0 && pwidth < JC.PanelWidth {
		h.refTitle.TextSize = JC.PanelTitleSizeSmall
		h.refSubtitle.TextSize = JC.PanelSubTitleSizeSmall
		h.refBottomText.TextSize = JC.PanelBottomTextSizeSmall
		h.refContent.TextSize = JC.PanelContentSizeSmall
	}

	pkt := JT.BP.GetDataByID(h.GetTag())
	if pkt == nil {
		return
	}

	if !JT.BP.ValidatePanel(pkt.Get()) {
		pkt.SetStatus(JC.STATE_BAD_CONFIG)
	}

	h.status = pkt.GetStatus()

	switch pkt.GetStatus() {
	case JC.STATE_ERROR:
		h.refTitle.Text = "Error loading data"
		h.refSubtitle.Hide()
		h.refBottomText.Hide()
		h.refContent.Hide()
		h.EnableClick()
		h.background.FillColor = JC.ErrorColor

	case JC.STATE_FETCHING_NEW:
		h.refTitle.Text = "Fetching Rates..."
		h.refSubtitle.Hide()
		h.refBottomText.Hide()
		h.refContent.Hide()
		h.EnableClick()
		h.background.FillColor = JC.PanelBG

	case JC.STATE_LOADING:
		h.refTitle.Text = "Loading..."
		h.refSubtitle.Hide()
		h.refBottomText.Hide()
		h.refContent.Hide()
		h.EnableClick()
		h.background.FillColor = JC.PanelBG

	case JC.STATE_BAD_CONFIG:
		h.refTitle.Text = "Invalid Panel"
		h.refSubtitle.Hide()
		h.refBottomText.Hide()
		h.refContent.Hide()
		h.background.FillColor = JC.ErrorColor
		h.EnableClick()

	default:

		h.refTitle.Text = JC.TruncateText(pkt.FormatTitle(), pwidth-20, h.refTitle.TextSize)
		h.refSubtitle.Text = JC.TruncateText(pkt.FormatSubtitle(), pwidth-20, h.refSubtitle.TextSize)
		h.refBottomText.Text = JC.TruncateText(pkt.FormatBottomText(), pwidth-20, h.refBottomText.TextSize)
		h.refContent.Text = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.refContent.TextSize)

		h.refSubtitle.Show()
		h.refBottomText.Show()
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
	if JM.AppStatus.IsDraggable() {
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

	if ActiveAction != nil {
		ActiveAction.HideTarget()
	}

	ActiveAction = h
}

func (h *PanelDisplay) HideTarget() {
	h.child.Hide()
	h.visible = false
	h.Refresh()
	ActiveAction = nil
}

func (h *PanelDisplay) Cursor() desktop.Cursor {
	if JM.AppStatus.IsDraggable() {
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

		rect, ok := JM.DragPlaceholder.(*canvas.Rectangle)
		if !ok {
			return
		}

		activeDragging = h
		h.dragging = true
		h.dragScroll = JM.AppLayout.OffsetY()

		p := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)

		rect.Move(p)
		canvas.Refresh(rect)

		// Try to show placeholder as soon as possible
		if p.X == rect.Position().X || p.Y == rect.Position().Y {
			rect, _ := JM.DragPlaceholder.(*canvas.Rectangle)
			rect.FillColor = JC.PanelPlaceholderBG
			canvas.Refresh(rect)
		}

		go func() {
			ticker := time.NewTicker(h.fps)
			defer ticker.Stop()

			rect, ok := JM.DragPlaceholder.(*canvas.Rectangle)

			if !ok {
				return
			}

			shown := rect.FillColor == JC.PanelPlaceholderBG
			posX := rect.Position().X
			posY := rect.Position().Y
			placeholderSize := rect.Size()
			edgeThreshold := placeholderSize.Height / 2
			scrollStep := float32(10)

			for h.dragging {
				<-ticker.C

				targetX := p.X + h.dragPosition.X - (placeholderSize.Width / 2)
				targetY := p.Y + h.dragPosition.Y - (placeholderSize.Height / 2)

				edgeTopY := JM.AppLayout.ContentTopY() - edgeThreshold
				edgeBottomY := JM.AppLayout.ContentBottomY() - edgeThreshold

				// Just in case the initial function failed to move and show
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
					JM.AppLayout.ScrollBy(-scrollStep)
				} else if posY > edgeBottomY {
					JM.AppLayout.ScrollBy(scrollStep)
				}
			}
		}()
	}
}

func (h *PanelDisplay) Dragged(ev *fyne.DragEvent) {
	if JM.AppStatus.IsDraggable() {
		h.PanelDrag(ev)
	} else {
		h.container.Dragged(ev)
	}
}

func (h *PanelDisplay) DragEnd() {
	if !JM.AppStatus.IsDraggable() {
		h.dragging = false
		if h.container != nil {
			h.container.DragEnd()
		}
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
	h.dragOffset.Y -= h.dragScroll - JM.AppLayout.OffsetY()

	h.snapToNearest()
}

func (h *PanelDisplay) snapToNearest() {

	// Convert target position to grid index
	targetIndex := h.findDropTargetIndex()
	if targetIndex == -1 {
		return
	}

	Grid.Objects = h.reorder(targetIndex)
	Grid.ForceRefresh()

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
			pdt := JT.BP.GetDataByID(uuid)
			if pdt != nil {
				nd = append(nd, pdt)
			}
		}
	}

	JT.BP.SetData(nd)

	return true
}
