package widgets

import (
	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
)

type dialogOverlaysLayout struct {
	background *dialogOverlays
	dialogBox  fyne.CanvasObject
	cHeight    float32
	cWidth     float32
}

func (l *dialogOverlaysLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Height == 0 || size.Width == 0 {
		return
	}

	if l.cWidth != size.Width {
		l.cHeight = 0
	}

	if JC.IsMobile {
		if l.cHeight == 0 {
			l.cHeight = size.Height
		}
	} else {
		l.cHeight = size.Height
	}

	l.cWidth = size.Width

	l.background.Resize(size)
	l.background.Move(fyne.NewPos(0, 0))
	l.background.Show()

	var dialogWidth float32
	switch {
	case l.cWidth <= 560:
		dialogWidth = l.cWidth - 10
	case l.cWidth > 560 && l.cWidth <= 1200:
		dialogWidth = l.cWidth * 0.8
	default:
		dialogWidth = 800
	}

	dialogHeight := l.dialogBox.MinSize().Height
	emptySpace := l.cHeight - dialogHeight
	posX := (l.cWidth - dialogWidth) / 2
	posY := emptySpace / 4

	if posY < 0 {
		posY = 0
	}

	l.dialogBox.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	l.dialogBox.Move(fyne.NewPos(posX, posY))
}

func (l *dialogOverlaysLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(300, 300)
}
