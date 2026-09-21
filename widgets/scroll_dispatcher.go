package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type scrollDispatcher struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	scroller *container.Scroll
}

func (s *scrollDispatcher) SetScroller(scroller *container.Scroll) {
	s.scroller = scroller
}

func (s *scrollDispatcher) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.content)
}

func (s *scrollDispatcher) Scrolled(ev *fyne.ScrollEvent) {
	if s.scroller == nil {
		return
	}

	if ev.Scrolled.DY != 0 {
		s.scroller.Scrolled(ev)
	}
}

func (s *scrollDispatcher) Dragged(ev *fyne.DragEvent) {
	if s.scroller == nil {
		return
	}

	// container.Scroll no longer exposes Dragged in fyne v2.8+.
	// Translate drag events into scroll events (invert Y to match
	// previous drag-to-scroll behavior).
	if ev.Dragged.DY != 0 || ev.Dragged.DX != 0 {
		s.scroller.Scrolled(&fyne.ScrollEvent{
			Scrolled: fyne.Delta{
				DX: ev.Dragged.DX,
				DY: -ev.Dragged.DY,
			},
		})
	}
}

func (s *scrollDispatcher) DragEnd() {
	// No explicit DragEnd on container.Scroll in fyne v2.8+; nothing to do.
}

func NewScrollDispatcher() *scrollDispatcher {
	s := &scrollDispatcher{
		content: canvas.NewRectangle(JC.UseTheme().GetColor(JC.ColorNameTransparent)),
	}
	s.ExtendBaseWidget(s)
	return s
}
