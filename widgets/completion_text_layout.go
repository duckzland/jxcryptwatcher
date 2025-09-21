package widgets

import (
	JC "jxwatcher/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type completionTextLayout struct {
	text       *canvas.Text
	separator  *canvas.Line
	background *canvas.Rectangle
	height     float32
}

func (r *completionTextLayout) Layout(size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	if !JC.IsMobile {
		r.background.Resize(size)
	}

	textHeight := r.text.TextSize
	yOffset := ((r.height - textHeight) / 2) - 4
	r.text.Move(fyne.NewPos(8, float32(yOffset)))

	text := ""
	if JC.IsMobile {
		text = JC.TruncateTextWithEstimation(r.text.Text, size.Width, r.text.TextSize)
	} else {
		text = JC.TruncateText(r.text.Text, size.Width, r.text.TextSize)
	}

	r.text.Text = text

	r.separator.Position1 = fyne.NewPos(0, r.height-1)
	r.separator.Position2 = fyne.NewPos(size.Width, r.height-1)
}

func (r *completionTextLayout) MinSize() fyne.Size {
	return fyne.NewSize(0, r.height-4)
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

func (r *completionTextLayout) Destroy() {}
