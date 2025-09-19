package apps

import "fyne.io/fyne/v2"

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
