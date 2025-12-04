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
}

func (s *completionText) CreateRenderer() fyne.WidgetRenderer {
	separator := canvas.NewLine(s.sepcolor)
	separator.StrokeWidth = 1

	if s.img == nil {
		s.rasterize()
	}

	return &completionTextLayout{
		parent:     s,
		text:       s.img,
		separator:  separator,
		background: s.background,
		height:     s.height,
	}
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

func (s *completionText) rasterize() {
	fs := JC.UseTheme().Size(JC.SizeCompletionText)

	dst, size := JC.RasterizeText(s.text, fyne.TextStyle{}, fs, JC.UseTheme().GetColor(theme.ColorNameForeground), s.cSize.Height/fs, s.cSize.Height-fs, JC.POS_LEFT, JC.POS_BOTTOM)
	if dst == nil || s.img == nil {
		return
	}

	s.img.Image = dst

	s.cSize = size
	s.img.SetMinSize(size)
	s.img.Resize(size)
	s.img.Refresh()
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

	s.ExtendBaseWidget(s)

	return s
}
