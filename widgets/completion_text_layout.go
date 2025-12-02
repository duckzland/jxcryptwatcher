package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"

	JC "jxwatcher/core"
)

type completionTextLayout struct {
	parent     *completionText
	text       *canvas.Image
	separator  *canvas.Line
	background *canvas.Rectangle
	height     float32
	width      float32
	cSize      fyne.Size
}

func (r *completionTextLayout) Layout(size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	if !JC.IsMobile {
		if r.background.Size() != size {
			r.background.Resize(size)
		}
	}

	textHeight := r.parent.textSize
	yOffset := ((r.height - textHeight) / 2) - theme.Padding()
	newPos := fyne.NewPos(8, float32(yOffset))

	if r.text.Position() != newPos {
		r.text.Move(newPos)
	}

	if r.width != size.Width {
		r.parent.SetText(r.parent.GetSource())
	}

	posY := r.height - 1
	pos1 := fyne.NewPos(0, posY)
	pos2 := fyne.NewPos(size.Width, posY)

	if r.separator.Position1 != pos1 {
		r.separator.Position1 = pos1
	}

	if r.separator.Position2 != pos2 {
		r.separator.Position2 = pos2
	}

	r.width = size.Width
}

func (r *completionTextLayout) MinSize() fyne.Size {
	if r.cSize.Height == 0 {
		r.cSize = fyne.NewSize(0, r.height-theme.Padding())
	}
	return r.cSize
}

func (r *completionTextLayout) Refresh() {
	if !JC.IsMobile {
		canvas.Refresh(r.background)
	}

	canvas.Refresh(r.text)
	canvas.Refresh(r.separator)
}

func (r *completionTextLayout) Objects() []fyne.CanvasObject {
	if JC.IsMobile {
		return []fyne.CanvasObject{r.text, r.separator}
	}

	return []fyne.CanvasObject{r.background, r.text, r.separator}
}

func (r *completionTextLayout) Destroy() {
	r.parent = nil
	r.text = nil
	r.separator = nil
	r.background = nil
}
