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
	cSize      fyne.Size
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

	if l.background.Size() != size {
		l.background.Resize(size)
	}

	if l.background.Position() != fyne.NewPos(0, 0) {
		l.background.Move(fyne.NewPos(0, 0))
	}

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
	maxHeight := l.cHeight * 0.95

	if JC.IsMobile {
		maxHeight = l.cHeight * 0.65
	}

	if dialogHeight > maxHeight {
		diff := dialogHeight - maxHeight
		if diff > 60 {
			dialogHeight = maxHeight
		}
	}

	emptySpace := l.cHeight - dialogHeight
	posX := (l.cWidth - dialogWidth) / 2
	posY := emptySpace / 4

	if posY < 0 {
		posY = 0
	}

	if JC.IsMobile && posY > 20 {
		posY = 20
	}

	dialogSize := fyne.NewSize(dialogWidth, dialogHeight)
	dialogPos := fyne.NewPos(posX, posY)

	if l.dialogBox.Size() != dialogSize {
		l.dialogBox.Resize(dialogSize)
	}
	if l.dialogBox.Position() != dialogPos {
		l.dialogBox.Move(dialogPos)
	}
}

func (l *dialogOverlaysLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.cSize.Height == 0 {
		l.cSize = fyne.NewSize(300, 300)
	}
	return l.cSize
}
