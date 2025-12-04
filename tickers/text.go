package tickers

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type tickerText struct {
	widget.BaseWidget
	index     int
	text      string
	color     color.Color
	textSize  float32
	textAlign fyne.TextAlign
	textStyle fyne.TextStyle
	cSize     fyne.Size
	img       *canvas.Image
}

func (s *tickerText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.img)
}

func (s *tickerText) MinSize() fyne.Size {
	if s.cSize != fyne.NewSize(0, 0) {
		return s.cSize
	}
	width := JC.MeasureText(s.text, s.textSize, s.textStyle)
	height := s.textSize * 1.35
	s.cSize = fyne.NewSize(width, height)
	return s.cSize
}

func (s *tickerText) Visible() bool {
	return s.BaseWidget.Visible() && s.text != JC.STRING_EMPTY
}

func (s *tickerText) GetText() string {
	return s.text
}

func (s *tickerText) SetTextSize(size float32) {
	if s.textSize == size {
		return
	}
	s.textSize = size
	s.cSize = fyne.NewSize(0, 0)
	s.rasterize()
	s.Refresh()
}

func (s *tickerText) SetText(t string) {
	if s.text == t {
		return
	}
	s.text = t
	s.cSize = fyne.NewSize(0, 0)

	if s.text == JC.STRING_EMPTY {
		s.Hide()
	} else {
		s.Show()
	}

	s.rasterize()
	s.Refresh()
}

func (s *tickerText) SetAlpha(a uint8) {
	r, g, b, _ := s.color.RGBA()
	s.SetColor(color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: a,
	})
}

func (s *tickerText) SetColor(col color.Color) {
	s.color = col
	s.rasterize()
	s.Refresh()
}

func (s *tickerText) rasterize() {

	dst, size := JC.RasterizeText(s.text, s.textStyle, s.textSize, s.color, 0.6, 2, JC.POS_CENTER, JC.POS_BOTTOM, JC.SCALE_BILINEAR)
	if dst == nil || s.img == nil {
		return
	}

	s.img.Image = dst

	s.cSize = size
	s.Resize(size)
	s.img.SetMinSize(size)
	s.img.Resize(size)
}

func NewTickerText(text string, col color.Color, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *tickerText {
	s := &tickerText{
		text:      text,
		color:     col,
		textSize:  size,
		textAlign: alignment,
		textStyle: style,
		img:       canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 0, 0))),
	}

	s.img.FillMode = canvas.ImageFillOriginal

	s.ExtendBaseWidget(s)

	return s
}
