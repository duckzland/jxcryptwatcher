package panels

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JM "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var activeDragging *panelDisplay = nil

type panelDisplay struct {
	widget.BaseWidget
	tag          string
	fps          time.Duration
	status       int
	visible      bool
	disabled     bool
	container    fyne.CanvasObject
	parent       *panelContainer
	dragScroll   float32
	dragPosition fyne.Position
	dragOffset   fyne.Position
	dragging     bool
	background   *canvas.Rectangle
	action       fyne.CanvasObject
	title        *panelText
	content      *panelText
	subtitle     *panelText
	bottomText   *panelText
}

func NewPanelDisplay(
	pdt JT.PanelData,
	onEdit func(pk string, uuid string),
	onDelete func(uuid string),
) *panelDisplay {

	uuid := JC.CreateUUID()
	pdt.SetID(uuid)
	str := pdt.GetData()

	tc := JC.ThemeColor(theme.ColorNameForeground)

	pl := &panelDisplayLayout{
		title:      NewPanelText("", tc, JC.ThemeSize(JC.SizePanelTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		subtitle:   NewPanelText("", tc, JC.ThemeSize(JC.SizePanelSubTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: false}),
		content:    NewPanelText("", tc, JC.ThemeSize(JC.SizePanelContent), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		background: canvas.NewRectangle(JC.ThemeColor(JC.ColorNamePanelBG)),
		bottomText: NewPanelText("", tc, JC.ThemeSize(JC.SizePanelBottomText), fyne.TextAlignCenter, fyne.TextStyle{Bold: false}),
	}

	pl.background.CornerRadius = JC.ThemeSize(JC.SizePanelBorderRadius)

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

	pd := &panelDisplay{
		tag: uuid,
		fps: 3 * time.Millisecond,
		container: container.New(
			pl,
			pl.background,
			pl.title,
			pl.content,
			pl.subtitle,
			pl.bottomText,
			pl.action,
		),
		action:     pl.action,
		visible:    false,
		disabled:   false,
		background: pl.background,
		title:      pl.title,
		content:    pl.content,
		subtitle:   pl.subtitle,
		bottomText: pl.bottomText,
	}

	if JC.IsMobile {
		pd.fps = 6 * time.Millisecond
	}

	pd.ExtendBaseWidget(pd)

	pd.action.Hide()

	pd.status = pdt.GetStatus()

	str.AddListener(binding.NewDataListener(pd.updateContent))

	JA.FadeInBackground(pd.background, 100*time.Millisecond, nil)

	return pd
}

func (h *panelDisplay) GetTag() string {
	return h.tag
}

func (h *panelDisplay) ShowTarget() {
	h.action.Show()
	h.visible = true
	h.Refresh()

	if h.parent.HasActiveAction() {
		h.parent.GetActiveAction().HideTarget()
	}

	h.parent.SetActiveAction(h)
}

func (h *panelDisplay) HideTarget() {
	h.action.Hide()
	h.visible = false
	h.Refresh()
	h.parent.ResetActiveAction()
}

func (h *panelDisplay) DisableClick() {
	h.disabled = true
}

func (h *panelDisplay) EnableClick() {
	h.disabled = false
}

func (h *panelDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.container)
}

func (h *panelDisplay) Tapped(event *fyne.PointEvent) {
	if JM.UseStatus().IsDraggable() {
		h.HideTarget()
		return
	}

	if h.visible && !h.action.Visible() {
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

func (h *panelDisplay) Cursor() desktop.Cursor {
	if JM.UseStatus().IsDraggable() {
		return desktop.PointerCursor
	}

	return desktop.DefaultCursor
}

func (h *panelDisplay) Dragged(ev *fyne.DragEvent) {
	if JM.UseStatus().IsDraggable() {
		h.panelDrag(ev)
	} else {
		h.parent.Dragged(ev)
	}
}

func (h *panelDisplay) DragEnd() {
	if !JM.UseStatus().IsDraggable() {
		h.dragging = false
		if h.parent != nil {
			h.parent.DragEnd()
		}
		return
	}

	// Call this early to cancel go routine
	activeDragging = nil
	h.dragging = false

	JM.UseLayout().SetPlaceholderColor(JC.ThemeColor(JC.ColorNameTransparent))
	JM.UseLayout().MovePlaceholder(fyne.NewPos(0, -JC.ThemeSize(JC.SizePanelHeight)))

	h.dragOffset = h.Position().Add(h.dragPosition)
	h.dragOffset.Y -= h.dragScroll - JM.UseLayout().OffsetY()

	h.snapToNearest()
}

func (h *panelDisplay) updateContent() {

	pkt := JT.UsePanelMaps().GetDataByID(h.GetTag())
	if pkt == nil {
		return
	}

	pwidth := h.Size().Width
	if pwidth != 0 && pwidth < JC.ThemeSize(JC.SizePanelWidth) {
		h.title.SetTextSize(JC.ThemeSize(JC.SizePanelTitleSmall))
		h.subtitle.SetTextSize(JC.ThemeSize(JC.SizePanelSubTitleSmall))
		h.bottomText.SetTextSize(JC.ThemeSize(JC.SizePanelBottomTextSmall))
		h.content.SetTextSize(JC.ThemeSize(JC.SizePanelContentSmall))
	}

	if !JT.UsePanelMaps().ValidatePanel(pkt.Get()) {
		pkt.SetStatus(JC.STATE_BAD_CONFIG)
	}

	title := ""
	subtitle := ""
	bottomText := ""
	content := ""
	background := JC.ThemeColor(JC.ColorNamePanelBG)
	status := h.status

	h.status = pkt.GetStatus()

	switch h.status {
	case JC.STATE_ERROR:
		title = "Error loading data"
		background = JC.ThemeColor(JC.ColorNameError)

	case JC.STATE_FETCHING_NEW:
		title = "Fetching Rates..."

	case JC.STATE_LOADING:
		title = "Loading..."

	case JC.STATE_BAD_CONFIG:
		title = "Invalid Panel"
		background = JC.ThemeColor(JC.ColorNameError)

	case JC.STATE_LOADED:
		if pkt.DidChange() {
			switch pkt.IsValueIncrease() {
			case JC.VALUE_INCREASE:
				background = JC.ThemeColor(JC.ColorNameGreen)

			case JC.VALUE_DECREASE:
				background = JC.ThemeColor(JC.ColorNameRed)
			}

		} else if pkt.IsOnInitialValue() {
			background = JC.ThemeColor(JC.ColorNamePanelBG)
		}

		title = JC.TruncateText(pkt.FormatTitle(), pwidth-20, h.title.GetText().TextSize, h.title.GetText().TextStyle)
		subtitle = JC.TruncateText(pkt.FormatSubtitle(), pwidth-20, h.subtitle.GetText().TextSize, h.subtitle.GetText().TextStyle)
		bottomText = JC.TruncateText(pkt.FormatBottomText(), pwidth-20, h.bottomText.GetText().TextSize, h.bottomText.GetText().TextStyle)
		content = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.content.GetText().TextSize, h.content.GetText().TextStyle)
	}

	h.EnableClick()

	h.title.SetText(title)
	h.subtitle.SetText(subtitle)
	h.bottomText.SetText(bottomText)
	h.content.SetText(content)

	if pkt.DidChange() {
		JA.StartFlashingText(h.content.GetText(), 50*time.Millisecond, h.content.GetText().Color, 1)
	}

	if h.background.FillColor != background {
		h.background.FillColor = background
		canvas.Refresh(h.background)
	}

	if h.status != status {
		h.Refresh()
	}
}

func (h *panelDisplay) panelDrag(ev *fyne.DragEvent) {
	if activeDragging != nil && activeDragging != h {
		activeDragging.DragEnd()
	}

	h.dragPosition = ev.Position

	if activeDragging == nil {

		activeDragging = h
		h.dragging = true
		h.dragScroll = JM.UseLayout().OffsetY()

		p := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		lm := JM.UseLayout()
		dp := JM.UseLayout().GetDragPlaceholder()

		lm.MovePlaceholder(p)

		dx := dp.Position().X
		dy := dp.Position().Y
		ds := dp.Size()

		// Try to show placeholder as soon as possible
		if p.X == dx || p.Y == dy {
			lm.SetPlaceholderColor(JC.ThemeColor(JC.ColorNamePanelPlaceholder))
		}

		go func() {
			ticker := time.NewTicker(h.fps)
			defer ticker.Stop()

			shown := lm.IsPlaceholderColor(JC.ThemeColor(JC.ColorNamePanelPlaceholder))
			posX := dx
			posY := dy
			placeholderSize := ds
			edgeThreshold := placeholderSize.Height / 2
			scrollStep := float32(10)

			for h.dragging {
				<-ticker.C

				targetX := p.X + h.dragPosition.X - (placeholderSize.Width / 2)
				targetY := p.Y + h.dragPosition.Y - (placeholderSize.Height / 2)

				edgeTopY := lm.ContentTopY() - edgeThreshold
				edgeBottomY := lm.ContentBottomY() - edgeThreshold

				// Just in case the initial function failed to move and show
				if !shown && (targetX == posX || targetY == posY) {
					fyne.Do(func() {
						lm.SetPlaceholderColor(JC.ThemeColor(JC.ColorNamePanelPlaceholder))
						shown = true
					})
				}

				// Move placeholder
				if posX != targetX || posY != targetY {
					posX = targetX
					posY = targetY
					fyne.Do(func() {
						lm.MovePlaceholder(fyne.NewPos(posX, posY))
					})
				}

				// Scroll when placeholder is half out of viewport
				if posY < edgeTopY {
					lm.ScrollBy(-scrollStep)
				} else if posY > edgeBottomY {
					lm.ScrollBy(scrollStep)
				}
			}
		}()
	}
}

func (h *panelDisplay) snapToNearest() {

	// Convert target position to grid index
	targetIndex := h.findDropTargetIndex()
	if targetIndex == -1 {
		return
	}

	UsePanelGrid().Objects = h.reorder(targetIndex)
	UsePanelGrid().ForceRefresh()

	go func() {
		JC.UseDebouncer().Call("panel_drag", 1000*time.Millisecond, func() {
			if h.syncPanelData() {
				if JT.SavePanels() {
					JC.Notify("Panels have been reordered and updated.")
				}
			}
		})
	}()
}

func (h *panelDisplay) findDropTargetIndex() int {

	JC.Logln(fmt.Sprintf("Dragging item - Position: (%.2f, %.2f)", h.dragOffset.X, h.dragOffset.Y))

	for i, zone := range dragDropZones {

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

func (h *panelDisplay) reorder(targetIndex int) []fyne.CanvasObject {
	panels := UsePanelGrid().Objects
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

func (h *panelDisplay) syncPanelData() bool {
	nd := []JT.PanelData{}

	for _, obj := range UsePanelGrid().Objects {
		if panel, ok := obj.(*panelDisplay); ok {
			uuid := panel.GetTag()
			pdt := JT.UsePanelMaps().GetDataByID(uuid)
			if pdt != nil {
				nd = append(nd, pdt)
			}
		}
	}

	JT.UsePanelMaps().SetData(nd)

	return true
}
