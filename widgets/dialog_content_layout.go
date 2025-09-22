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

	if l.background.Size() != size {
		l.background.Resize(size)
	}

	if l.background.Position() != fyne.NewPos(0, 0) {
		l.background.Move(fyne.NewPos(0, 0))
	}

	x := l.padding
	y := l.padding
	w := size.Width - 2*l.padding

	titleSize := l.title.MinSize()
	titleTargetSize := fyne.NewSize(w, titleSize.Height)
	titleTargetPos := fyne.NewPos(x, y)

	if l.title.Size() != titleTargetSize {
		l.title.Resize(titleTargetSize)
	}

	if l.title.Position() != titleTargetPos {
		l.title.Move(titleTargetPos)
	}

	y += titleSize.Height + l.padding

	for _, top := range l.topContent {
		topSize := top.MinSize()
		targetSize := fyne.NewSize(w, topSize.Height)
		targetPos := fyne.NewPos(x, y)

		if top.Size() != targetSize {
			top.Resize(targetSize)
		}

		if top.Position() != targetPos {
			top.Move(targetPos)
		}

		y += topSize.Height + l.padding
	}

	formSize := l.form.MinSize()
	formTargetSize := fyne.NewSize(w, formSize.Height)
	formTargetPos := fyne.NewPos(x, y)

	if l.form.Size() != formTargetSize {
		l.form.Resize(formTargetSize)
	}

	if l.form.Position() != formTargetPos {
		l.form.Move(formTargetPos)
	}

	y += formSize.Height + l.padding

	for _, bottom := range l.bottomContent {
		bottomSize := bottom.MinSize()
		targetSize := fyne.NewSize(w, bottomSize.Height)
		targetPos := fyne.NewPos(x, y)

		if bottom.Size() != targetSize {
			bottom.Resize(targetSize)
		}

		if bottom.Position() != targetPos {
			bottom.Move(targetPos)
		}

		y += bottomSize.Height + l.padding
	}

	buttonSize := l.buttons.MinSize()
	buttonX := x + (w-buttonSize.Width)/2
	buttonTargetPos := fyne.NewPos(buttonX, y)

	if l.buttons.Size() != buttonSize {
		l.buttons.Resize(buttonSize)
	}

	if l.buttons.Position() != buttonTargetPos {
		l.buttons.Move(buttonTargetPos)
	}

	y += buttonSize.Height + l.padding
}

func (l *dialogContentLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := l.title.MinSize().Height +
		l.form.MinSize().Height +
		l.buttons.MinSize().Height +
		4*l.padding

	return fyne.NewSize(0, h)
}
