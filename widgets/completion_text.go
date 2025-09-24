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
	parent     *completionList
	text       string
	label      *canvas.Text
	height     float32
	hovered    bool
	background *canvas.Rectangle
	bgcolor    color.Color
	sepcolor   color.Color
	traColor   color.Color
}

func NewCompletionText(height float32, parent *completionList) *completionText {
	s := &completionText{
		label:    canvas.NewText("", JC.UseTheme().GetColor(theme.ColorNameForeground)),
		height:   height,
		parent:   parent,
		bgcolor:  JC.UseTheme().GetColor(theme.ColorNameHover),
		sepcolor: JC.UseTheme().GetColor(theme.ColorNameSeparator),
		traColor: JC.UseTheme().GetColor(JC.ColorNameTransparent),
	}

	if !JC.IsMobile {
		s.background = canvas.NewRectangle(s.traColor)
	}

	s.label.TextSize = JC.UseTheme().Size(JC.SizeCompletionText)
	s.label.Alignment = fyne.TextAlignLeading
	s.ExtendBaseWidget(s)
	return s
}

func (s *completionText) CreateRenderer() fyne.WidgetRenderer {
	separator := canvas.NewLine(s.sepcolor)
	separator.StrokeWidth = 1

	return &completionTextLayout{
		text:       s.label,
		separator:  separator,
		background: s.background,
		height:     s.height,
	}
}

func (s *completionText) SetText(t string) {
	if s.text == t {
		return
	}
	s.text = t
	s.label.Text = t
	canvas.Refresh(s.label)
}

func (s *completionText) SetIndex(i int) {
	s.index = i
}

func (s *completionText) SetParent(p *completionList) {
	s.parent = p
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
