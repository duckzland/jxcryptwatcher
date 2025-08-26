package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type EmptyClickablePanelLayout struct{}

func (p *EmptyClickablePanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}

	bg := objects[0]
	title := objects[1]
	content := objects[2]

	// Stretch background to fill parent
	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	// Get sizes of title and content
	titleSize := title.MinSize()
	contentSize := content.MinSize()

	// Total height of centered items
	totalHeight := titleSize.Height + contentSize.Height
	startY := (size.Height - totalHeight) / 2

	// Center title
	title.Move(fyne.NewPos((size.Width-titleSize.Width)/2, startY))
	title.Resize(titleSize)

	// Center content below title
	content.Move(fyne.NewPos((size.Width-contentSize.Width)/2, startY+titleSize.Height-10))
	content.Resize(contentSize)
}

func (p *EmptyClickablePanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	if len(objects) >= 3 {
		title := objects[1]
		content := objects[2]

		titleSize := title.MinSize()
		contentSize := content.MinSize()

		width = fyne.Max(titleSize.Width, contentSize.Width)
		height = titleSize.Height + contentSize.Height
	}

	return fyne.NewSize(width, height)
}

type EmptyClickablePanel struct {
	widget.BaseWidget
	OnTapped func()
	hovered  bool
}

func NewEmptyClickablePanel(onTap func()) *EmptyClickablePanel {
	p := &EmptyClickablePanel{
		OnTapped: onTap,
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *EmptyClickablePanel) CreateRenderer() fyne.WidgetRenderer {
	plus := canvas.NewText("+", JC.TextColor)
	plus.TextStyle = fyne.TextStyle{Bold: true}
	plus.Alignment = fyne.TextAlignCenter
	plus.TextSize = 64

	label := canvas.NewText("Add Panel", JC.TextColor)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 20

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	content := container.New(&EmptyClickablePanelLayout{},
		background,
		plus,
		label,
	)

	return widget.NewSimpleRenderer(content)
}

func (p *EmptyClickablePanel) Tapped(_ *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped()
	}
}

func (p *EmptyClickablePanel) TappedSecondary(_ *fyne.PointEvent) {}

func (p *EmptyClickablePanel) MouseIn(_ *desktop.MouseEvent) {
	p.hovered = true
	p.Refresh()
}

func (p *EmptyClickablePanel) MouseOut() {
	p.hovered = false
	p.Refresh()
}

func (p *EmptyClickablePanel) MouseMoved(_ *desktop.MouseEvent) {}

func (p *EmptyClickablePanel) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
