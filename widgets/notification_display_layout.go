package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type notificationDisplayLayout struct {
	padding   float32
	text      *canvas.Image
	container *notificationDisplay
	cSize     fyne.Size
}

func (r *notificationDisplayLayout) Layout(size fyne.Size) {
	r.MinSize()

	w := size.Width - r.padding*2
	h := r.cSize.Height
	if h <= 0 {
		h = 20
	}

	if r.text != nil {
		x := (size.Width-r.text.Size().Width)/2 + r.padding
		y := (40 - h + 2) / 2

		pos := fyne.NewPos(x, y)
		if r.text.Position() != pos {
			r.text.Move(pos)
		}
	}

	if r.container != nil {
		r.container.pSize = fyne.NewSize(w, h)
	}
}

func (r *notificationDisplayLayout) MinSize() fyne.Size {
	if r.cSize.Height == 0 {
		r.cSize = fyne.NewSize(200, 20)
	}
	return r.cSize
}

func (r *notificationDisplayLayout) Refresh() {
	fyne.Do(func() {
		canvas.Refresh(r.text)
	})
}

func (r *notificationDisplayLayout) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *notificationDisplayLayout) Destroy() {}
