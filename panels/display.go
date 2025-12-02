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

type PanelDisplay interface {
	GetTag() string
	Visible() bool
	Show()
	Hide()
	ShowTarget()
	HideTarget()
	Destroy()
}

type panelDisplay struct {
	widget.BaseWidget
	tag           string
	fps           time.Duration
	status        int
	shown         int
	actionVisible bool
	dragScroll    float32
	dragPosition  fyne.Position
	dragOffset    fyne.Position
	dragging      bool
	container     *fyne.Container
	background    *canvas.Rectangle
	action        *panelAction
	title         *panelText
	content       *panelText
	subtitle      *panelText
	bottomText    *panelText
	onEdit        func()
	onDelete      func()
}

func (h *panelDisplay) GetTag() string {
	return h.tag
}

func (h *panelDisplay) ShowTarget() {
	h.createAction()
	h.actionVisible = true
	h.Refresh()

	if UsePanelGrid().HasActiveAction() {
		UsePanelGrid().GetActiveAction().HideTarget()
	}

	UsePanelGrid().SetActiveAction(h)
}

func (h *panelDisplay) HideTarget() {
	h.removeAction()
	h.actionVisible = false
	h.Refresh()
	UsePanelGrid().ResetActiveAction()
}

func (h *panelDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.container)
}

func (h *panelDisplay) Show() {

	tc := JC.UseTheme().GetColor(theme.ColorNameForeground)

	h.background = canvas.NewRectangle(JC.UseTheme().GetColor(JC.ColorNamePanelBG))
	h.background.CornerRadius = JC.UseTheme().Size(JC.SizePanelBorderRadius)

	h.title = NewPanelText(JC.STRING_EMPTY, tc, JC.UseTheme().Size(JC.SizePanelTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	h.subtitle = NewPanelText(JC.STRING_EMPTY, tc, JC.UseTheme().Size(JC.SizePanelSubTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	h.content = NewPanelText(JC.STRING_EMPTY, tc, JC.UseTheme().Size(JC.SizePanelContent), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	h.bottomText = NewPanelText(JC.STRING_EMPTY, tc, JC.UseTheme().Size(JC.SizePanelBottomText), fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	h.container.Layout.(*panelDisplayLayout).RemoveAll()
	h.container.Layout.(*panelDisplayLayout).SetContent(h.background, h.title, h.subtitle, h.content, h.bottomText, nil)
	h.container.Objects = []fyne.CanvasObject{
		h.background,
		h.title,
		h.subtitle,
		h.content,
		h.bottomText,
	}

	if h.actionVisible {
		h.createAction()
	}

	h.BaseWidget.Show()

	h.updateContent()
	h.Refresh()

	if h.shown == 0 {
		h.shown++
		if !JC.IsMobile {
			if JM.UseLayout().UseScroll().Offset.X == 0 {
				JA.StartFadeInBackground(h.tag, h.background, 300*time.Millisecond, nil, false)
			}
		}
	}
}

func (h *panelDisplay) Hide() {
	h.BaseWidget.Hide()

	JA.StopFlashingText(h.tag)
	JA.StopFadeInBackground(h.tag)
	JA.StopFadeOutBackground(h.tag)

	h.background = nil
	h.title = nil
	h.subtitle = nil
	h.content = nil
	h.bottomText = nil

	h.removeAction()

	h.container.Layout.(*panelDisplayLayout).RemoveAll()
	h.container.RemoveAll()
}

func (h *panelDisplay) Tapped(event *fyne.PointEvent) {

	if JM.UseStatus().IsDraggable() {
		return
	}

	if h.actionVisible && !h.action.Visible() {
		h.ShowTarget()
		return
	}

	if activeDragging != nil {
		return
	}

	if h.actionVisible {
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
		UsePanelGrid().Dragged(ev)
	}
}

func (h *panelDisplay) DragEnd() {
	if !JM.UseStatus().IsDraggable() {
		h.dragging = false
		UsePanelGrid().DragEnd()
		return
	}

	// Call this early to cancel go routine
	activeDragging = nil
	h.dragging = false

	JM.UseLayout().UsePlaceholder().SetColor(JC.UseTheme().GetColor(JC.ColorNameTransparent))
	JM.UseLayout().UsePlaceholder().Move(fyne.NewPos(0, -JC.UseTheme().Size(JC.SizePanelHeight)))

	h.dragOffset = h.Position().Add(h.dragPosition)
	h.dragOffset.Y -= h.dragScroll - JM.UseLayout().UseScroll().Offset.Y

	h.snapToNearest()
}

func (h *panelDisplay) Refresh() {
	if h.container != nil {
		h.container.Refresh()
	}
}

func (h *panelDisplay) Destroy() {
	h.Hide()

	h.container = nil
	h.tag = JC.STRING_EMPTY
	h.status = 0
	h.shown = 0
	h.fps = 0
	h.actionVisible = false
	h.dragScroll = 0
	h.dragPosition = fyne.Position{}
	h.dragOffset = fyne.Position{}
	h.dragging = false
}

func (h *panelDisplay) updateContent() {

	if !h.Visible() {
		return
	}

	pkt := JT.UsePanelMaps().GetDataByID(h.GetTag())
	if pkt == nil {
		return
	}

	pwidth := h.Size().Width
	if pwidth != 0 {
		if pwidth < JC.UseTheme().Size(JC.SizePanelWidth) {
			h.title.SetTextSize(JC.UseTheme().Size(JC.SizePanelTitleSmall))
			h.subtitle.SetTextSize(JC.UseTheme().Size(JC.SizePanelSubTitleSmall))
			h.bottomText.SetTextSize(JC.UseTheme().Size(JC.SizePanelBottomTextSmall))
			h.content.SetTextSize(JC.UseTheme().Size(JC.SizePanelContentSmall))
		} else {
			h.title.SetTextSize(JC.UseTheme().Size(JC.SizePanelTitle))
			h.subtitle.SetTextSize(JC.UseTheme().Size(JC.SizePanelSubTitle))
			h.bottomText.SetTextSize(JC.UseTheme().Size(JC.SizePanelBottomText))
			h.content.SetTextSize(JC.UseTheme().Size(JC.SizePanelContent))
		}
	}

	if !JT.UsePanelMaps().ValidatePanel(pkt.Get()) {
		pkt.SetStatus(JC.STATE_BAD_CONFIG)
	}

	title := JC.STRING_EMPTY
	subtitle := JC.STRING_EMPTY
	bottomText := JC.STRING_EMPTY
	content := JC.STRING_EMPTY
	background := JC.UseTheme().GetColor(JC.ColorNamePanelBG)
	status := h.status

	h.status = pkt.GetStatus()

	switch h.status {
	case JC.STATE_ERROR:
		title = "Error loading data"
		background = JC.UseTheme().GetColor(JC.ColorNameError)

	case JC.STATE_FETCHING_NEW:
		title = "Fetching Rates..."

	case JC.STATE_LOADING:
		title = "Loading..."

	case JC.STATE_BAD_CONFIG:
		title = "Invalid Panel"
		background = JC.UseTheme().GetColor(JC.ColorNameError)

	case JC.STATE_LOADED:
		if pkt.DidChange() {
			switch pkt.IsValueIncrease() {
			case JC.VALUE_INCREASE:
				background = JC.UseTheme().GetColor(JC.ColorNameGreen)

			case JC.VALUE_DECREASE:
				background = JC.UseTheme().GetColor(JC.ColorNameRed)
			}

		} else if pkt.IsOnInitialValue() {
			background = JC.UseTheme().GetColor(JC.ColorNamePanelBG)
		}

		title = JC.TruncateText(pkt.FormatTitle(), pwidth-20, h.title.textSize, h.title.textStyle)
		subtitle = JC.TruncateText(pkt.FormatSubtitle(), pwidth-20, h.subtitle.textSize, h.subtitle.textStyle)
		bottomText = JC.TruncateText(pkt.FormatBottomText(), pwidth-20, h.bottomText.textSize, h.bottomText.textStyle)
		content = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.content.textSize, h.content.textStyle)
	}

	h.title.SetText(title)
	h.subtitle.SetText(subtitle)
	h.bottomText.SetText(bottomText)
	h.content.SetText(content)

	if pkt.DidChange() {
		if h.Visible() {
			JA.StartFlashingText(h.tag, h.content, 50*time.Millisecond, JC.UseTheme().GetColor(theme.ColorNameForeground), 1)
		}
	}

	if h.background.FillColor != background {
		h.background.FillColor = background
		if h.Visible() {
			if JC.IsMobile {
				canvas.Refresh(h.background)
			} else {
				JA.StartFadeInBackground(h.tag, h.background, 300*time.Millisecond, nil, false)
			}
		}
	}

	if h.status != status {
		if h.Visible() {
			h.Refresh()
		}
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
		h.dragScroll = JM.UseLayout().UseScroll().Offset.Y

		p := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		lm := JM.UseLayout()
		dp := JM.UseLayout().UsePlaceholder()
		dc := JC.UseTheme().GetColor(JC.ColorNamePanelPlaceholder)

		dp.Move(p)

		dx := dp.Position().X
		dy := dp.Position().Y
		ds := dp.Size()

		// Try to show placeholder as soon as possible
		if p.X == dx || p.Y == dy {
			dp.SetColor(dc)
		}

		ctY := lm.UseScroll().Position().Y
		cbY := ctY + lm.UseScroll().Size().Height

		go func() {
			ticker := time.NewTicker(h.fps)
			defer ticker.Stop()

			shown := dp.IsColor(dc)
			posX := dx
			posY := dy
			placeholderSize := ds
			edgeThreshold := placeholderSize.Height / 2
			scrollStep := float32(3)

			for h.dragging {
				<-ticker.C

				targetX := p.X + h.dragPosition.X - (placeholderSize.Width / 2)
				targetY := p.Y + h.dragPosition.Y - (placeholderSize.Height / 2)

				edgeTopY := ctY - edgeThreshold
				edgeBottomY := cbY - edgeThreshold

				// Just in case the initial function failed to move and show
				if !shown && (targetX == posX || targetY == posY) {
					fyne.Do(func() {
						dp.SetColor(dc)
						shown = true
					})
				}

				// Move placeholder
				if posX != targetX || posY != targetY {
					posX = targetX
					posY = targetY
					fyne.Do(func() {
						dp.Move(fyne.NewPos(posX, posY))
					})
				}

				if posY > edgeBottomY || posY < edgeTopY {
					deltaY := float32(0)

					if posY > edgeBottomY {
						deltaY = posY - edgeBottomY
						deltaY = -fyne.Min(deltaY, scrollStep)
					} else {
						deltaY = edgeTopY - posY
						deltaY = fyne.Min(deltaY, scrollStep)
					}

					fyne.Do(func() {
						lm.UseScroll().Scrolled(&fyne.ScrollEvent{
							Scrolled: fyne.Delta{DX: 0, DY: deltaY},
						})
					})
				}
			}
		}()
	}
}

func (h *panelDisplay) snapToNearest() {

	// Convert target position to grid index
	targetIndex := h.findTargetIndex()
	if targetIndex == -1 {
		return
	}

	UsePanelGrid().Objects = h.reorder(targetIndex)
	UsePanelGrid().ForceRefresh()

	JC.UseDebouncer().Call("panel_drag", 1000*time.Millisecond, func() {
		if h.syncData() {
			if JT.SavePanels() {
				JC.Notify(JC.NotifyPanelsHaveBeenReorderedAndUpdated)
			}
		}
	})
}

func (h *panelDisplay) findTargetIndex() int {

	JC.Logln(fmt.Sprintf("Dragging item - Position: (%.2f, %.2f)", h.dragOffset.X, h.dragOffset.Y))

	for i, zone := range dragDropZones {

		JC.Logln(fmt.Sprintf("Checking panel %d — Bounds: [X: %.2f–%.2f, Y: %.2f–%.2f]", i, zone.left, zone.right, zone.top, zone.bottom))

		if h.dragOffset.X >= zone.left &&
			h.dragOffset.X <= zone.right &&
			h.dragOffset.Y >= zone.top &&
			h.dragOffset.Y <= zone.bottom {

			if zone.uuid == h.GetTag() {
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

	if !h.Visible() {
		h.Show()
	}

	if !JC.IsMobile {
		JA.StartFadeInBackground(h.tag, h.background, 300*time.Millisecond, nil, false)
	}

	return result
}

func (h *panelDisplay) syncData() bool {
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

func (h *panelDisplay) createAction() {
	if h.action == nil {
		h.action = NewPanelAction(h.onEdit, h.onDelete)
		h.container.Layout.(*panelDisplayLayout).action = h.action
		h.container.Objects = append(h.container.Objects, h.action)
		h.action.Show()
	}
}

func (h *panelDisplay) removeAction() {
	if h.action != nil {
		h.action.Hide()
		h.container.Layout.(*panelDisplayLayout).action = nil
		h.container.Remove(h.action)
		h.action = nil
	}
}

func NewPanelDisplay(pdt JT.PanelData, onEdit func(pk string, uuid string), onDelete func(uuid string)) *panelDisplay {

	uuid := JC.CreateUUID()
	pdt.SetID(uuid)
	pv := pdt.UseData()
	ps := pdt.UseStatus()

	pd := &panelDisplay{
		tag:           uuid,
		fps:           3 * time.Millisecond,
		container:     container.New(&panelDisplayLayout{}),
		shown:         0,
		actionVisible: false,
	}

	if onEdit != nil {
		pd.onEdit = func() {
			pdt := JT.UsePanelMaps().GetDataByID(pd.GetTag())
			pv := pdt.UseData()
			dynpk, _ := pv.Get()

			onEdit(dynpk, pd.GetTag())
		}
	}

	if onDelete != nil {
		pd.onDelete = func() {
			JA.StartFadeOutBackground(pd.tag, pd.background, 300*time.Millisecond, func() {
				onDelete(pd.GetTag())
			}, false)

			pd.HideTarget()
		}
	}

	if JC.IsMobile {
		pd.fps = 6 * time.Millisecond
	}

	pd.ExtendBaseWidget(pd)

	pd.status = pdt.GetStatus()

	pv.AddListener(binding.NewDataListener(pd.updateContent))
	ps.AddListener(binding.NewDataListener(pd.updateContent))

	pd.Hide()

	return pd
}
