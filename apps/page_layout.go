package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

type pageLayout struct {
	background *canvas.Rectangle
	icon       *fyne.Container
	content    *canvas.Text
	cSize      fyne.Size
}

func (p *pageLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	if p.background != nil {
		p.background.Resize(size)
		p.background.Move(fyne.NewPos(0, 0))
	}

	contentSize := fyne.NewSize(0, 0)
	iconSize := fyne.NewSize(64, 64)
	totalHeight := float32(0)

	if p.icon != nil {
		totalHeight += iconSize.Height
	}

	if p.content != nil {
		textWidth := JC.MeasureText(p.content.Text, p.content.TextSize, p.content.TextStyle)
		contentSize = fyne.NewSize(textWidth+20, p.content.TextSize*2)
		totalHeight += contentSize.Height
	}

	startY := (size.Height - totalHeight) / 2

	if p.icon != nil {
		p.icon.Move(fyne.NewPos((size.Width-iconSize.Width)/2, startY))
		p.icon.Resize(iconSize)

		if len(p.icon.Objects) > 0 {
			innerIcon := p.icon.Objects[0]
			innerIcon.Resize(iconSize)
			innerIcon.Move(fyne.NewPos(0, 0))
		}

		startY += iconSize.Height
	}

	if p.content != nil {
		p.content.Move(fyne.NewPos((size.Width-contentSize.Width)/2, startY))
		p.content.Resize(contentSize)
	}
}

func (p *pageLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if p.cSize.Height == 0 {
		width := float32(0)
		height := float32(0)

		if p.icon != nil {
			width = 64
			height += 64
		}

		if p.content != nil {
			textWidth := JC.MeasureText(p.content.Text, p.content.TextSize, p.content.TextStyle)
			if width < textWidth {
				width = textWidth
			}
			height += p.content.TextSize * 2
		}

		p.cSize = fyne.NewSize(width, height)
	}

	return p.cSize
}
