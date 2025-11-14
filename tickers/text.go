package tickers

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type tickerText struct {
	widget.BaseWidget
	index int
	text  string
	label *canvas.Text
	cSize fyne.Size
}

func (s *tickerText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

func (s *tickerText) MinSize() fyne.Size {
	if s.cSize != fyne.NewSize(0, 0) {
		return s.cSize
	}

	width := JC.MeasureText(s.text, s.label.TextSize, s.label.TextStyle)
	height := s.label.TextSize * 1.35
	s.cSize = fyne.NewSize(width, height)

	return s.cSize
}

func (s *tickerText) GetText() *canvas.Text {
	return s.label
}

func (s *tickerText) SetTextSize(size float32) {
	s.label.TextSize = size
	s.cSize = fyne.NewSize(0, 0)
}

func (s *tickerText) SetText(t string) {
	if s.text == t {
		return
	}

	if t == JC.STRING_EMPTY {
		s.label.Hide()
	} else {
		s.label.Show()
	}

	s.text = t
	s.label.Text = t
	s.cSize = fyne.NewSize(0, 0)
	canvas.Refresh(s.label)
}

func (s *tickerText) Visible() bool {
	return s.label.Visible()
}

func NewTickerText(text string, color color.Color, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *tickerText {
	s := &tickerText{
		label: canvas.NewText(text, color),
		cSize: fyne.NewSize(0, 0),
	}

	s.label.TextSize = size
	s.label.Alignment = alignment
	s.label.TextStyle = style
	s.ExtendBaseWidget(s)
	return s
}
