package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type pageLayout struct {
	background *canvas.Rectangle
	icon       *fyne.Container
	content    *canvas.Text
}

func (p *pageLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	if p.background != nil {
		p.background.Resize(size)
		p.background.Move(fyne.NewPos(0, 0))
	}

	contentSize := p.content.MinSize()
	iconSize := fyne.NewSize(64, 64)
	totalHeight := float32(0)

	if p.icon != nil {
		totalHeight += iconSize.Height
	}

	if p.content != nil {
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
	width := float32(0)
	height := float32(0)

	if p.icon != nil {
		ic := p.icon.MinSize()
		width += ic.Width
		height += ic.Height
	}

	if p.content != nil {
		co := p.content.MinSize()
		if width < co.Width {
			width = co.Width
		}
		height += co.Height

	}

	return fyne.NewSize(width, height)
}
