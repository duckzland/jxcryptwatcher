package tickers

import (
	JC "jxwatcher/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type tickerLayout struct {
	background *canvas.Rectangle
	title      *canvas.Text
	content    *canvas.Text
	status     *canvas.Text
}

func (tl *tickerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	spacer := float32(-2)

	tl.background.Resize(size)
	tl.background.Move(fyne.NewPos(0, 0))

	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{tl.title, tl.content, tl.status} {
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
}

func (tl *tickerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(JC.TickerWidth, JC.TickerHeight)
}
