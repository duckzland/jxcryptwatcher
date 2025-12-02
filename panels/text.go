package panels

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

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
	r, g, b, _ := p.color.RGBA()
	p.SetColor(color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: a,
	})
}

func (p *panelText) SetColor(col color.Color) {
	p.color = col
	p.rasterize()
	p.Refresh()
}

func (p *panelText) rasterize() {
	scale := JC.Window.Canvas().Scale()

	face := JC.UseTheme().GetFontFace(p.textStyle, p.textSize)
	if face == nil {
		return
	}

	adv := font.MeasureString(face, p.text)
	textW := max(adv.Round(), 1)
	padding := p.textSize * 0.35
	if padding > 4 {
		padding = 4
	}
	height := p.textSize + padding
	width := int(float32(textW) * scale)

	buf := image.NewRGBA(image.Rect(0, 0, width, int(height)))

	startX := (width - textW) / 2

	d := &font.Drawer{
		Dst:  buf,
		Src:  image.NewUniform(p.color),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(startX << 6),
			Y: fixed.Int26_6(int(height-padding) << 6),
		},
	}
	d.DrawString(p.text)

	if p.img == nil {
		p.img = canvas.NewImageFromImage(buf)
	} else {
		p.img.Image = buf
	}

	p.img.FillMode = canvas.ImageFillOriginal
	size := fyne.NewSize(float32(buf.Bounds().Dx()), height)
	p.cSize = size
	p.img.SetMinSize(size)
	p.img.Resize(size)
}

func NewPanelText(text string, col color.Color, size float32, alignment fyne.TextAlign, style fyne.TextStyle) *panelText {
	s := &panelText{
		text:      text,
		color:     col,
		textSize:  size,
		textAlign: alignment,
		textStyle: style,
	}

	s.ExtendBaseWidget(s)

	return s
}
