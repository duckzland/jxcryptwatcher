package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type notificationDisplayLayout struct {
	text      *canvas.Text
	container *notificationDisplay
}

func (r *notificationDisplayLayout) Layout(size fyne.Size) {
	p := r.container.padding
	pos := fyne.NewPos(p, 0)
	if r.text.Position() != pos {
		r.text.Move(fyne.NewPos(p, 0))
	}

	pz := fyne.NewSize(size.Width-2*p, size.Height)
	if r.text.Size() != pz {
		r.text.Resize(pz)
	}
}

func (r *notificationDisplayLayout) MinSize() fyne.Size {
	return fyne.NewSize(200, 20)
}

func (r *notificationDisplayLayout) Refresh() {
	r.text.Refresh()
}

func (r *notificationDisplayLayout) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *notificationDisplayLayout) Destroy() {}
