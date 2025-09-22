package tickers

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type tickerText struct {
	widget.BaseWidget
	index int
	text  string
	label *canvas.Text
}

func NewTickerText(text string, color color.RGBA, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *tickerText {
	s := &tickerText{
		label: canvas.NewText(text, color),
	}

	s.label.TextSize = size
	s.label.Alignment = alignment
	s.label.TextStyle = style
	s.ExtendBaseWidget(s)
	return s
}

func (s *tickerText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

func (s *tickerText) GetText() *canvas.Text {
	return s.label
}

func (s *tickerText) SetTextSize(size float32) {
	s.label.TextSize = size
}

func (s *tickerText) SetText(t string) {
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
	canvas.Refresh(s.label)
}

func (s *tickerText) Visible() bool {
	return s.label.Visible()
}
