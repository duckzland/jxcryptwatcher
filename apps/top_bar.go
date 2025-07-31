package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

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

func NewTopBar(
	onCryptosRefresh func(),
	onRatesRefresh func(),
	onSettingSave func(),
	onAddNewPanel func(),
) fyne.CanvasObject {

	topBg := canvas.NewRectangle(JC.PanelBG)
	topBg.CornerRadius = 4

	return container.New(
		&TopBarLayout{
			fixedWidth: JC.ActionBtnWidth,
			spacer:     JC.ActionBtnGap,
		},

		container.NewStack(
			topBg,
			JW.NewNotificationDisplayWidget(JC.UpdateStatusChan),
		),

		// Refresh ticker data
		JW.NewHoverCursorIconButton("", theme.ViewRestoreIcon(), "Refresh ticker data", func() {
			if onCryptosRefresh != nil {
				go onCryptosRefresh()
			}
		}),

		// Refresh exchange rates
		JW.NewHoverCursorIconButton("", theme.ViewRefreshIcon(), "Update rates from exchange", func() {
			if onRatesRefresh != nil {
				go onRatesRefresh()
			}
		}),

		// Open settings
		JW.NewHoverCursorIconButton("", theme.SettingsIcon(), "Open settings", func() {
			if onSettingSave != nil {
				go onSettingSave()
			}
		}),

		// Add new panel
		JW.NewHoverCursorIconButton("", theme.ContentAddIcon(), "Add new panel", func() {
			if onAddNewPanel != nil {
				go onAddNewPanel()
			}
		}),
	)
}
