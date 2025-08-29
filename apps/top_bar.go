package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JW "jxwatcher/widgets"
)

type TopBarLayout struct {
	fixedWidth float32
	spacer     float32
	rows       int
}

func (s *TopBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	count := len(objects)
	if count == 0 {
		return
	}

	// First object fills the rest of the space
	remaining := size.Width - (s.fixedWidth+s.spacer)*float32(count-1)

	s.rows = 1
	xh := size.Height / 2

	if remaining < 500 {

		s.rows = 2

		// Layout objects
		curPos := float32(0)
		y := float32(0)
		for i, obj := range objects {
			var w float32
			w = s.fixedWidth
			y = 0

			switch i {
			case 0:
				w = size.Width
				y = xh + s.spacer
				curPos = 0

			case 1:
				curPos += remaining/2 + s.spacer

			default:
				curPos += w + s.spacer
			}

			obj.Resize(fyne.NewSize(w, xh))
			obj.Move(fyne.NewPos(curPos, y))
		}

	} else {
		// Layout objects
		curPos := float32(0)
		for i, obj := range objects {
			var w float32
			if i == 0 {
				w = remaining
			} else {
				w = s.fixedWidth
			}

			obj.Resize(fyne.NewSize(w, size.Height))
			obj.Move(fyne.NewPos(curPos, 0))

			curPos += w + s.spacer
		}
	}
}

func (s *TopBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var maxHeight float32
	for _, obj := range objects {
		h := obj.MinSize().Height
		if h > maxHeight {
			maxHeight = h
		}
	}
	width := s.fixedWidth*float32(len(objects)-1) + 400
	if s.rows > 1 {
		maxHeight = float32(s.rows)*maxHeight + s.spacer
	}

	return fyne.NewSize(width, maxHeight)
}

func NewTopBar() fyne.CanvasObject {

	topBg := canvas.NewRectangle(JC.PanelBG)
	topBg.CornerRadius = 4

	JC.NotificationContainer = topBg

	return container.New(
		&TopBarLayout{
			fixedWidth: JC.ActionBtnWidth,
			spacer:     JC.ActionBtnGap,
		},
		container.NewStack(
			topBg,
			JW.NewNotificationDisplayWidget(JC.UpdateStatusChan),
		),
		AppActionManager.GetButton("refresh_cryptos"),
		AppActionManager.GetButton("refresh_rates"),
		AppActionManager.GetButton("open_settings"),
		AppActionManager.GetButton("toggle_drag"),
		AppActionManager.GetButton("add_panel"),
	)
}
