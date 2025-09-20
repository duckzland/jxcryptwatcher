package panels

import (
	"time"

	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
)

type panelContainerLayout struct {
	container *panelContainer
}

func (r *panelContainerLayout) Layout(size fyne.Size) {
	r.container.layout.Layout(r.container.Objects, size)
}

func (r *panelContainerLayout) MinSize() fyne.Size {
	return r.container.layout.MinSize(r.container.Objects)
}

func (r *panelContainerLayout) Refresh() {
	JC.MainDebouncer.Call("panel_container_refresh", 10*time.Millisecond, func() {
		fyne.Do(func() {
			r.Layout(r.container.Size())
		})
	})
}

func (r *panelContainerLayout) Objects() []fyne.CanvasObject {
	return r.container.Objects
}

func (r *panelContainerLayout) Destroy() {
	r.container.dragging = false
}
