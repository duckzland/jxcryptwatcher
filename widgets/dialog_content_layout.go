package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type dialogContentLayout struct {
	background    *canvas.Rectangle
	content       *container.Scroll
	form          fyne.CanvasObject
	buttons       fyne.CanvasObject
	title         fyne.CanvasObject
	topContent    []*fyne.Container
	bottomContent []*fyne.Container
	padding       float32
	cWidth        float32
	cHeight       float32
	cSize         fyne.Size
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

	// Calculate available height for the form
	ch := size.Height - (l.title.MinSize().Height + l.padding) - (l.buttons.MinSize().Height + l.padding) - 2*l.padding
	for _, x := range l.bottomContent {
		ch -= (x.MinSize().Height + l.padding)
	}

	for _, x := range l.topContent {
		ch -= (x.MinSize().Height + l.padding)
	}

	// Title
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

	// Top content
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

	// Main content
	contentTargetSize := fyne.NewSize(w, ch)
	contentTargetPos := fyne.NewPos(x, y)

	if l.content.Size() != contentTargetSize {
		l.content.Resize(contentTargetSize)
	}

	if l.content.Position() != contentTargetPos {
		l.content.Move(contentTargetPos)
	}

	y += contentTargetSize.Height + l.padding

	// Bottom content
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

	// Buttons
	buttonSize := l.buttons.MinSize()
	buttonX := x + (w-buttonSize.Width)/2
	buttonTargetPos := fyne.NewPos(buttonX, y)

	if l.buttons.Size() != buttonSize {
		l.buttons.Resize(buttonSize)
	}

	if l.buttons.Position() != buttonTargetPos {
		l.buttons.Move(buttonTargetPos)
	}
}

func (l *dialogContentLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.cHeight == 0 {
		l.cHeight = l.title.MinSize().Height +
			l.form.MinSize().Height +
			l.buttons.MinSize().Height +
			4*l.padding

		l.cSize = fyne.NewSize(0, l.cHeight)
	}

	return l.cSize
}
