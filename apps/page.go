package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type AppPageLayout struct{}

func (p *AppPageLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	bg := objects[0]
	var icon fyne.CanvasObject
	var content fyne.CanvasObject

	if len(objects) == 2 {
		content = objects[1]
	} else {
		icon = objects[1]
		content = objects[2]
	}

	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	contentSize := content.MinSize()
	iconSize := fyne.NewSize(64, 64)

	var totalHeight float32
	if icon != nil {
		totalHeight = iconSize.Height + contentSize.Height
	} else {
		totalHeight = contentSize.Height
	}

	startY := (size.Height - totalHeight) / 2

	if icon != nil {
		icon.Move(fyne.NewPos((size.Width-iconSize.Width)/2, startY))
		icon.Resize(iconSize)

		if c, ok := icon.(*fyne.Container); ok && len(c.Objects) > 0 {
			innerIcon := c.Objects[0]
			innerIcon.Resize(iconSize)
			innerIcon.Move(fyne.NewPos(0, 0))
		}

		startY += iconSize.Height
	}

	// Position content
	content.Move(fyne.NewPos((size.Width-contentSize.Width)/2, startY))
	content.Resize(contentSize)
}

func (p *AppPageLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	if len(objects) >= 3 {
		icon := objects[1]
		content := objects[2]

		iconSize := icon.MinSize()
		contentSize := content.MinSize()

		width = fyne.Max(iconSize.Width, contentSize.Width)
		height = iconSize.Height + contentSize.Height
	}

	return fyne.NewSize(width, height)
}

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
