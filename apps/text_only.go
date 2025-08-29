package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type TextOnlyPanelLayout struct{}

func (p *TextOnlyPanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	bg := objects[0]
	content := objects[1]

	// Stretch background to fill parent
	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	// Get sizes of title and content
	contentSize := content.MinSize()

	// Total height of centered items
	totalHeight := contentSize.Height
	startY := (size.Height - totalHeight) / 2

	// Center content below title
	content.Move(fyne.NewPos((size.Width-contentSize.Width)/2, startY))
	content.Resize(contentSize)
}

func (p *TextOnlyPanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	if len(objects) >= 1 {
		content := objects[1]
		contentSize := content.MinSize()

		width = contentSize.Width
		height = contentSize.Height
	}

	return fyne.NewSize(width, height)
}

type TextOnlyPanel struct {
	widget.BaseWidget
	text string
}

func NewTextOnlyPanel(text string) *TextOnlyPanel {
	p := &TextOnlyPanel{
		text: text,
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *TextOnlyPanel) CreateRenderer() fyne.WidgetRenderer {
	label := canvas.NewText(p.text, JC.TextColor)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 20

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	content := container.New(&TextOnlyPanelLayout{},
		background,
		label,
	)

	return widget.NewSimpleRenderer(content)
}
