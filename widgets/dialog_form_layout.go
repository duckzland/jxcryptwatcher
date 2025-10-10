package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type dialogFormLayout struct {
	form       *widget.Form
	container  *container.Scroll
	dispatcher *scrollDispatcher
	cSize      fyne.Size
}

func (l *dialogFormLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width == 0 || size.Height == 0 {
		return
	}

	formSize := l.form.MinSize()
	formTargetSize := fyne.NewSize(size.Width, formSize.Height)
	formTargetPos := fyne.NewPos(0, 0)

	if formSize != l.cSize {
		l.cSize = formSize
	}

	if formTargetSize.Height > l.container.Size().Height {
		formTargetSize.Width -= 18
	}

	if l.form.Position() != formTargetPos {
		l.form.Move(formTargetPos)
	}

	if l.form.Size() != formTargetSize {
		l.form.Resize(formTargetSize)
	}

	if l.dispatcher.Size() != formTargetSize {
		l.dispatcher.Resize(formTargetSize)
	}

	if l.dispatcher.Position() != formTargetPos {
		l.dispatcher.Move(formTargetPos)
	}

}

func (l *dialogFormLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.cSize.Height == 0 {
		l.cSize = fyne.NewSize(0, l.form.MinSize().Height)
	}

	return l.cSize
}
