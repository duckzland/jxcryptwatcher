package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type panelDisplayLayout struct {
	background *canvas.Rectangle
	title      *canvas.Text
	subtitle   *canvas.Text
	content    *canvas.Text
	bottomText *canvas.Text
	action     *panelAction
}

func (pl *panelDisplayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width <= 0 && size.Height <= 0 {
		return
	}

	if len(objects) < 5 {
		return
	}

	spacer := float32(-2)

	pl.background.Resize(size)
	pl.background.Move(fyne.NewPos(0, 0))

	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{pl.title, pl.content, pl.subtitle, pl.bottomText} {
		if obj.Visible() && obj.MinSize().Height > 0 {
			centerItems = append(centerItems, obj)
		}
	}

	var totalHeight float32
	for _, obj := range centerItems {
		totalHeight += obj.MinSize().Height
	}

	totalHeight += spacer * float32(len(centerItems)-1)

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	for _, obj := range centerItems {
		objSize := obj.MinSize()
		obj.Move(fyne.NewPos((size.Width-objSize.Width)/2, currentY))
		obj.Resize(objSize)
		currentY += objSize.Height + spacer
	}

	actionSize := pl.action.MinSize()
	pl.action.Move(fyne.NewPos(size.Width-actionSize.Width, 0))
	pl.action.Resize(actionSize)
}

func (pl *panelDisplayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	for _, obj := range objects[1:5] {
		if obj.Visible() && obj.MinSize().Height > 0 {
			size := obj.MinSize()
			if size.Width > width {
				width = size.Width
			}
			height += size.Height
		}
	}

	return fyne.NewSize(width, height)
}
