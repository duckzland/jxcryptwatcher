package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type AppPage struct {
	widget.BaseWidget
	OnTapped func()
	hovered  bool
	icon     *fyne.Resource
	content  string
}

func NewAppPage(icon *fyne.Resource, content string, onTap func()) *AppPage {
	p := &AppPage{
		icon:     icon,
		content:  content,
		OnTapped: onTap,
	}

	p.ExtendBaseWidget(p)
	return p
}

func (p *AppPage) CreateRenderer() fyne.WidgetRenderer {
	var objects []fyne.CanvasObject

	content := canvas.NewText(p.content, JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextSize = 20

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	objects = append(objects, background)

	if p.icon != nil {
		objects = append(objects, container.NewWithoutLayout(widget.NewIcon(*p.icon)))
	}

	objects = append(objects, content)

	containerContent := container.New(&AppPageLayout{}, objects...)

	return widget.NewSimpleRenderer(containerContent)
}

func (p *AppPage) Tapped(_ *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped()
	}
}

func (p *AppPage) TappedSecondary(_ *fyne.PointEvent) {}

func (p *AppPage) MouseIn(_ *desktop.MouseEvent) {
	p.hovered = true
	p.Refresh()
}

func (p *AppPage) MouseOut() {
	p.hovered = false
	p.Refresh()
}

func (p *AppPage) MouseMoved(_ *desktop.MouseEvent) {}

func (p *AppPage) Cursor() desktop.Cursor {
	if p.OnTapped != nil {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}
