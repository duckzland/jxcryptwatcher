package tickers

import (
	"fyne.io/fyne/v2"
)

type tickerContainerLayout struct {
	container *tickerContainer
}

func (r *tickerContainerLayout) Layout(size fyne.Size) {
	r.container.layout.Layout(r.container.Objects, size)
}

func (r *tickerContainerLayout) MinSize() fyne.Size {
	return r.container.layout.MinSize(r.container.Objects)
}

func (r *tickerContainerLayout) Refresh() {
	r.Layout(r.container.Size())
}

func (r *tickerContainerLayout) Objects() []fyne.CanvasObject {
	return r.container.Objects
}

func (r *tickerContainerLayout) Destroy() {
}
