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

func (r *notificationDisplayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	nSize := size
	if r.cSize.Height > 0 {
		nSize.Height = r.cSize.Height
	} else {
		nSize.Height = 20
	}

	if r.text != nil {
		nSize.Width -= r.padding * 2
		pos := fyne.NewPos(r.padding, (40-nSize.Height+2)/2)
		if r.text.Position() != pos {
			r.text.Move(pos)
		}

		pz := fyne.NewSize(size.Width-2*r.padding, nSize.Height+2)
		if r.text.Size() != pz {
			r.text.Resize(pz)
		}
	}

	if r.container != nil {
		r.container.pSize = nSize
	}
}

func (r *notificationDisplayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
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
