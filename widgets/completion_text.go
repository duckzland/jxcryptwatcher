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
}

func NewCompletionText(height float32, parent *completionList) *completionText {
	s := &completionText{
		label:    canvas.NewText("", JC.MainTheme.Color(theme.ColorNameForeground, theme.VariantDark)),
		height:   height,
		parent:   parent,
		bgcolor:  JC.MainTheme.Color(theme.ColorNameHover, theme.VariantDark),
		sepcolor: JC.MainTheme.Color(theme.ColorNameSeparator, theme.VariantDark),
	}

	if !JC.IsMobile {
		s.background = canvas.NewRectangle(JC.Transparent)
	}

	s.label.TextSize = JC.CompletionTextSize
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
	s.background.FillColor = JC.Transparent
	canvas.Refresh(s)
}

func (s *completionText) MouseMoved(*desktop.MouseEvent) {}
