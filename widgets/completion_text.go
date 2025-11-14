package widgets

import (
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
	parent     *completionList
	label      *canvas.Text
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

	return &completionTextLayout{
		parent:     s,
		text:       s.label,
		separator:  separator,
		background: s.background,
		height:     s.height,
	}
}

func (s *completionText) GetText() string {
	return s.text
}

func (s *completionText) SetText(t string) {
	if s.text == t {
		return
	}
	size := s.Size()

	s.text = t
	s.label.Text = JC.TruncateText(s.GetText(), size.Width, s.label.TextSize, s.label.TextStyle)
	canvas.Refresh(s.label)
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
	canvas.Refresh(s.label)
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
	canvas.Refresh(s)
}

func (s *completionText) MouseMoved(*desktop.MouseEvent) {}

func NewCompletionText(width float32, height float32, parent *completionList) *completionText {
	s := &completionText{
		label:    canvas.NewText(JC.STRING_EMPTY, JC.UseTheme().GetColor(theme.ColorNameForeground)),
		width:    width,
		height:   height,
		parent:   parent,
		bgcolor:  JC.UseTheme().GetColor(theme.ColorNameHover),
		sepcolor: JC.UseTheme().GetColor(theme.ColorNameSeparator),
		traColor: JC.UseTheme().GetColor(JC.ColorNameTransparent),
		cSize:    fyne.NewSize(width, height-theme.Padding()),
	}

	if !JC.IsMobile {
		s.background = canvas.NewRectangle(s.traColor)
	}

	s.label.TextSize = JC.UseTheme().Size(JC.SizeCompletionText)
	s.label.Alignment = fyne.TextAlignLeading
	s.ExtendBaseWidget(s)
	return s
}
