package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type staticPage struct {
	widget.BaseWidget
	OnTapped func()
	hovered  bool
	icon     *fyne.Resource
	content  string
}

func NewAppPage(icon *fyne.Resource, content string, onTap func()) *staticPage {
	p := &staticPage{
		icon:     icon,
		content:  content,
		OnTapped: onTap,
	}

	p.ExtendBaseWidget(p)

	return p
}

func (p *staticPage) CreateRenderer() fyne.WidgetRenderer {
	var objects []fyne.CanvasObject

	layout := &pageLayout{
		background: canvas.NewRectangle(JC.PanelBG),
		content:    canvas.NewText(p.content, JC.MainTheme.Color(theme.ColorNameForeground, theme.VariantDark)),
	}

	layout.content.Alignment = fyne.TextAlignCenter
	layout.content.TextSize = 20

	layout.background.SetMinSize(fyne.NewSize(100, 100))
	layout.background.CornerRadius = JC.PanelBorderRadius

	objects = append(objects, layout.background)

	if p.icon != nil {
		layout.icon = container.NewWithoutLayout(widget.NewIcon(*p.icon))
		objects = append(objects, layout.icon)
	}

	objects = append(objects, layout.content)

	return widget.NewSimpleRenderer(container.New(layout, objects...))
}

func (p *staticPage) Tapped(_ *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped()
	}
}

func (p *staticPage) TappedSecondary(_ *fyne.PointEvent) {}

func (p *staticPage) MouseIn(_ *desktop.MouseEvent) {
	p.hovered = true
	p.Refresh()
}

func (p *staticPage) MouseOut() {
	p.hovered = false
	p.Refresh()
}

func (p *staticPage) MouseMoved(_ *desktop.MouseEvent) {}

func (p *staticPage) Cursor() desktop.Cursor {
	if p.OnTapped != nil {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}
