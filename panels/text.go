package panels

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type panelText struct {
	widget.BaseWidget
	index int
	text  string
	label *canvas.Text
}

func (s *panelText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

func (s *panelText) GetText() *canvas.Text {
	return s.label
}

func (s *panelText) SetTextSize(size float32) {
	s.label.TextSize = size
}

func (s *panelText) SetText(t string) {
	if s.text == t {
		return
	}

	if t == "" {
		s.label.Hide()
	} else {
		s.label.Show()
	}

	s.text = t
	s.label.Text = t
	if s.Visible() {
		canvas.Refresh(s.label)
	}
}

func (s *panelText) Visible() bool {
	return s.BaseWidget.Visible() && s.label.Visible()
}

func NewPanelText(text string, color color.Color, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *panelText {
	s := &panelText{
		label: canvas.NewText(text, color),
	}

	s.label.TextSize = size
	s.label.Alignment = alignment
	s.label.TextStyle = style
	s.ExtendBaseWidget(s)
	return s
}
