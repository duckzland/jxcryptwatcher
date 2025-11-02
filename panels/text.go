package panels

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type panelText struct {
	widget.BaseWidget
	index int
	text  string
	label *canvas.Text
	cSize fyne.Size
}

func (s *panelText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

func (s *panelText) MinSize() fyne.Size {
	if s.cSize != fyne.NewSize(0, 0) {
		return s.cSize
	}

	width := JC.MeasureText(s.text, s.label.TextSize, s.label.TextStyle)
	height := s.label.TextSize * 1.35
	s.cSize = fyne.NewSize(width, height)

	return s.cSize
}

func (s *panelText) GetText() *canvas.Text {
	return s.label
}

func (s *panelText) SetTextSize(size float32) {
	s.label.TextSize = size
	s.cSize = fyne.NewSize(0, 0)
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
	s.cSize = fyne.NewSize(0, 0)

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
		cSize: fyne.NewSize(0, 0),
	}

	s.label.TextSize = size
	s.label.Alignment = alignment
	s.label.TextStyle = style
	s.ExtendBaseWidget(s)
	return s
}
