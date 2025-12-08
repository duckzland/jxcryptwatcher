package widgets

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type completionText struct {
	widget.BaseWidget
	index      int
	text       string
	source     string
	textSize   float32
	textStyle  fyne.TextStyle
	parent     *completionList
	img        *canvas.Image
	background *canvas.Rectangle
	width      float32
	height     float32
	hovered    bool
	bgcolor    color.Color
	sepcolor   color.Color
	traColor   color.Color
	cSize      fyne.Size
	renderer   *completionTextLayout
}

func (s *completionText) CreateRenderer() fyne.WidgetRenderer {
	separator := canvas.NewLine(s.sepcolor)
	separator.StrokeWidth = 1

	if s.img == nil {
		s.rasterize()
	}
	r := &completionTextLayout{
		parent:     s,
		text:       s.img,
		separator:  separator,
		background: s.background,
		height:     s.height,
	}
	s.renderer = r
	return r

}

func (s *completionText) GetText() string {
	return s.text
}

func (s *completionText) GetSource() string {
	return s.source
}

func (s *completionText) SetText(t string) {

	s.source = t
	maxWidth := s.Size().Width
	txt := JC.TruncateText(t, maxWidth, s.textSize, s.textStyle)

	if s.text == txt {
		return
	}

	s.text = txt

	s.rasterize()
	s.Refresh()

}

func (s *completionText) SetIndex(i int) {
	s.index = i
}

func (s *completionText) SetParent(p *completionList) {
	s.parent = p
}

func (p *completionText) MinSize() fyne.Size {
	return p.cSize
}

func (s *completionText) Tapped(_ *fyne.PointEvent) {
	if s.parent.IsDragging() {
		return
	}

	if s.parent != nil && s.index >= 0 {
		s.parent.OnSelected(s.index)
	}
}

func (s *completionText) MouseIn(*desktop.MouseEvent) {

	if JC.IsMobile {
		return
	}

	if s.parent.IsDragging() {
		s.MouseOut()
		return
	}

	if s.hovered == true {
		return
	}

	s.hovered = true
	s.background.FillColor = s.bgcolor
	canvas.Refresh(s.img)
}

func (s *completionText) MouseOut() {

	if JC.IsMobile {
		return
	}

	if s.hovered == false {
		return
	}

	s.hovered = false
	s.background.FillColor = s.traColor
	canvas.Refresh(s.img)
}

func (s *completionText) MouseMoved(*desktop.MouseEvent) {}

func (c *completionText) Destroy() {

	if c == nil {
		return
	}

	if c.img != nil {
		if c.img.Image != nil {
			c.img.Image = nil
		}
		c.img = nil
	}

	c.background = nil

	c.text = JC.STRING_EMPTY
	c.source = JC.STRING_EMPTY

	c.textSize = 0
	c.textStyle = fyne.TextStyle{}
	c.width = 0
	c.height = 0
	c.cSize = fyne.Size{}

	c.bgcolor = nil
	c.sepcolor = nil
	c.traColor = nil

	c.hovered = false
	c.parent = nil

	if c.renderer != nil {
		c.renderer.Destroy()
		c.renderer = nil
	}

	c.ExtendBaseWidget(nil)
}

func (s *completionText) rasterize() {

	if s.img == nil {
		return
	}

	fs := JC.UseTheme().Size(JC.SizeCompletionText)

	current, _ := s.img.Image.(*image.NRGBA)
	dst := JC.RasterizeText(current, s.text, fyne.TextStyle{Bold: false}, fs, fyne.TextAlignLeading, JC.UseTheme().GetColor(theme.ColorNameForeground))
	if dst == nil {
		return
	}

	size := fyne.NewSize(float32(dst.Bounds().Dx()), float32(dst.Bounds().Dy()))
	s.img.Resize(size)
	s.img.Image = dst
	s.img.Refresh()

	dst = nil
}

func NewCompletionText(width float32, height float32, parent *completionList) *completionText {
	s := &completionText{
		width:     width,
		height:    height,
		parent:    parent,
		bgcolor:   JC.UseTheme().GetColor(theme.ColorNameHover),
		sepcolor:  JC.UseTheme().GetColor(theme.ColorNameSeparator),
		traColor:  JC.UseTheme().GetColor(JC.ColorNameTransparent),
		cSize:     fyne.NewSize(width, height-theme.Padding()),
		textSize:  JC.UseTheme().Size(JC.SizeCompletionText),
		textStyle: fyne.TextStyle{Bold: false},
		img:       canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 0, 0))),
	}

	if !JC.IsMobile {
		s.background = canvas.NewRectangle(s.traColor)
	}

	s.img.FillMode = canvas.ImageFillOriginal
	s.img.ScaleMode = canvas.ImageScaleFastest

	s.ExtendBaseWidget(s)

	return s
}
