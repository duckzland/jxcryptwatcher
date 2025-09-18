package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type SelectableText struct {
	widget.BaseWidget
	index  int
	parent *navigableList
	text   string
	label  *canvas.Text
}

func NewSelectableText() *SelectableText {
	s := &SelectableText{
		label: canvas.NewText("", JC.TextColor),
	}
	s.label.TextSize = 14
	s.label.Alignment = fyne.TextAlignLeading
	s.ExtendBaseWidget(s)
	return s
}

func (s *SelectableText) CreateRenderer() fyne.WidgetRenderer {
	return &selectableTextRenderer{text: s.label}
}

func (s *SelectableText) SetText(t string) {
	if s.text == t {
		return
	}

	s.text = t
	if s.label != nil {
		s.label.Text = t
		canvas.Refresh(s.label)
	}
}

func (s *SelectableText) SetIndex(i int) {
	s.index = i
}

func (s *SelectableText) SetParent(p *navigableList) {
	s.parent = p
}

func (s *SelectableText) Tapped(_ *fyne.PointEvent) {
	if s.parent != nil && s.index >= 0 {
		if s.parent.OnSelected != nil {
			s.parent.OnSelected(s.index)
		}
	}
}

type selectableTextRenderer struct {
	text *canvas.Text
}

func (r *selectableTextRenderer) Layout(size fyne.Size) {
	textHeight := r.text.TextSize + 2
	yOffset := (36 - textHeight) / 2
	r.text.Move(fyne.NewPos(8, float32(yOffset)))
}

func (r *selectableTextRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 36)
}

func (r *selectableTextRenderer) Refresh() {
}

func (r *selectableTextRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *selectableTextRenderer) Destroy() {
}
