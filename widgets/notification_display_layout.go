package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type notificationDisplayLayout struct {
	text      *canvas.Text
	container *NotificationDisplay
}

func (r *notificationDisplayLayout) Layout(size fyne.Size) {
	p := r.container.padding
	r.text.Move(fyne.NewPos(p, 0))
	r.text.Resize(fyne.NewSize(size.Width-2*p, size.Height))
}

func (r *notificationDisplayLayout) MinSize() fyne.Size {
	min := r.text.MinSize()
	p := r.container.padding
	return fyne.NewSize(min.Width+2*p, min.Height)
}

func (r *notificationDisplayLayout) Refresh() {
	r.text.Refresh()
}

func (r *notificationDisplayLayout) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *notificationDisplayLayout) Destroy() {}
