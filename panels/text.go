package panels

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type panelText struct {
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

func (p *panelText) CreateRenderer() fyne.WidgetRenderer {
	if p.img == nil {
		p.rasterize()
	}
	return widget.NewSimpleRenderer(p.img)
}

func (p *panelText) MinSize() fyne.Size {
	if p.cSize != fyne.NewSize(0, 0) {
		return p.cSize
	}
	width := JC.MeasureText(p.text, p.textSize, p.textStyle)
	height := p.textSize * 1.35
	p.cSize = fyne.NewSize(width, height)
	return p.cSize
}

func (p *panelText) Visible() bool {
	return p.BaseWidget.Visible() && p.text != JC.STRING_EMPTY
}

func (p *panelText) GetText() string {
	return p.text
}

func (p *panelText) SetTextSize(size float32) {
	if p.textSize == size {
		return
	}
	p.textSize = size
	p.cSize = fyne.NewSize(0, 0)
	p.rasterize()
	p.Refresh()
}

func (p *panelText) SetText(t string) {
	if p.text == t {
		return
	}
	p.text = t
	p.cSize = fyne.NewSize(0, 0)

	if p.text == JC.STRING_EMPTY {
		p.Hide()
	} else {
		p.Show()
	}

	p.rasterize()
	p.Refresh()
}

func (p *panelText) SetAlpha(a uint8) {
	if p.img == nil {
		return
	}
	JC.SetImageAlpha(p.img.Image.(*image.NRGBA), a)
	p.img.Refresh()
}

func (p *panelText) SetColor(col color.Color) {
	if p.img == nil {
		return
	}
	JC.SetImageColor(p.img.Image.(*image.NRGBA), col)
	p.img.Refresh()
}

func (p *panelText) Destroy() {
	if p == nil {
		return
	}

	if p.img != nil {
		if p.img.Image != nil {
			p.img.Image = nil
		}
		p.img = nil
	}

	p.text = JC.STRING_EMPTY
	p.color = nil
	p.textSize = 0
	p.textAlign = fyne.TextAlignLeading
	p.textStyle = fyne.TextStyle{}
	p.cSize = fyne.Size{}

	p.ExtendBaseWidget(nil)
}

func (p *panelText) rasterize() {

	if p.img == nil {
		return
	}

	current, _ := p.img.Image.(*image.NRGBA)
	dst := JC.RasterizeText(current, p.text, p.textStyle, p.textSize, p.textAlign, p.color)
	if dst == nil {
		return
	}

	size := fyne.NewSize(float32(dst.Bounds().Dx()), float32(dst.Bounds().Dy()))
	p.img.Resize(size)

	p.img.Image = dst

	p.cSize = size
	p.Resize(size)
	p.img.Refresh()

	dst = nil
}

func NewPanelText(text string, col color.Color, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *panelText {
	s := &panelText{
		text:      text,
		color:     col,
		textSize:  size,
		textAlign: alignment,
		textStyle: style,
		img:       canvas.NewImageFromImage(image.NewNRGBA(image.Rect(0, 0, 0, 0))),
	}

	s.img.FillMode = canvas.ImageFillOriginal
	s.img.ScaleMode = canvas.ImageScaleFastest

	s.ExtendBaseWidget(s)

	return s
}
