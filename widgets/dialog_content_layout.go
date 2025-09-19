package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type dialogContentLayout struct {
	background    *canvas.Rectangle
	title         fyne.CanvasObject
	topContent    []*fyne.Container
	form          fyne.CanvasObject
	buttons       fyne.CanvasObject
	bottomContent []*fyne.Container
	padding       float32
	cWidth        float32
	cHeight       float32
}

func (l *dialogContentLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width == 0 || size.Height == 0 {
		return
	}

	if l.cWidth == size.Width && l.cHeight == size.Height {
		return
	}

	l.cWidth = size.Width
	l.cHeight = size.Height

	l.background.Resize(size)
	l.background.Move(fyne.NewPos(0, 0))

	x := l.padding
	y := l.padding
	w := size.Width - 2*l.padding

	titleSize := l.title.MinSize()
	l.title.Resize(fyne.NewSize(w, titleSize.Height))
	l.title.Move(fyne.NewPos(x, y))
	y += titleSize.Height + l.padding

	for _, top := range l.topContent {
		topSize := top.MinSize()
		top.Resize(fyne.NewSize(w, topSize.Height))
		top.Move(fyne.NewPos(x, y))
		y += topSize.Height + l.padding
	}

	formSize := l.form.MinSize()
	l.form.Resize(fyne.NewSize(w, formSize.Height))
	l.form.Move(fyne.NewPos(x, y))
	y += formSize.Height + l.padding

	for _, bottom := range l.bottomContent {
		bottomSize := bottom.MinSize()
		bottom.Resize(fyne.NewSize(w, bottomSize.Height))
		bottom.Move(fyne.NewPos(x, y))
		y += bottomSize.Height + l.padding
	}

	buttonSize := l.buttons.MinSize()
	buttonX := x + (w-buttonSize.Width)/2
	l.buttons.Resize(buttonSize)
	l.buttons.Move(fyne.NewPos(buttonX, y))
	y += buttonSize.Height + l.padding
}

func (l *dialogContentLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := l.title.MinSize().Height +
		l.form.MinSize().Height +
		l.buttons.MinSize().Height +
		4*l.padding

	return fyne.NewSize(0, h)
}
