package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type ActionNeededClickablePanelLayout struct{}

func (p *ActionNeededClickablePanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}

	bg := objects[0]
	icon := objects[1]
	content := objects[2]

	// Stretch background to fill parent
	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	// Get sizes of icon and content
	iconSize := fyne.NewSize(64, 64)
	contentSize := content.MinSize()

	// Total height of centered items
	totalHeight := iconSize.Height + contentSize.Height
	startY := (size.Height - totalHeight) / 2

	// Center icon
	icon.Move(fyne.NewPos((size.Width-iconSize.Width)/2, startY))
	icon.Resize(iconSize)

	if c, ok := icon.(*fyne.Container); ok && len(c.Objects) > 0 {
		icon := c.Objects[0]
		icon.Resize(iconSize)
		icon.Move(fyne.NewPos(0, 0))
	}

	// Center content below icon
	content.Move(fyne.NewPos((size.Width-contentSize.Width)/2, startY+iconSize.Height))
	content.Resize(contentSize)
}

func (p *ActionNeededClickablePanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
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

type ActionNeededClickablePanel struct {
	widget.BaseWidget
	OnTapped func()
	hovered  bool
	icon     fyne.Resource
	content  string
}

func NewActionNeededClickablePanel(icon fyne.Resource, content string, onTap func()) *ActionNeededClickablePanel {
	p := &ActionNeededClickablePanel{
		icon:     icon,
		content:  content,
		OnTapped: onTap,
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *ActionNeededClickablePanel) CreateRenderer() fyne.WidgetRenderer {

	icon := container.NewWithoutLayout(widget.NewIcon(p.icon))

	label := canvas.NewText(p.content, JC.TextColor)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 20

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	content := container.New(&ActionNeededClickablePanelLayout{},
		background,
		icon,
		label,
	)

	return widget.NewSimpleRenderer(content)
}

func (p *ActionNeededClickablePanel) Tapped(_ *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped()
	}
}

func (p *ActionNeededClickablePanel) TappedSecondary(_ *fyne.PointEvent) {}

func (p *ActionNeededClickablePanel) MouseIn(_ *desktop.MouseEvent) {
	p.hovered = true
	p.Refresh()
}

func (p *ActionNeededClickablePanel) MouseOut() {
	p.hovered = false
	p.Refresh()
}

func (p *ActionNeededClickablePanel) MouseMoved(_ *desktop.MouseEvent) {}

func (p *ActionNeededClickablePanel) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
