package tickers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var tickerDisplayLayoutCachedSize fyne.Size

type tickerLayout struct {
	background *canvas.Rectangle
	title      *tickerText
	content    *tickerText
	status     *tickerText
}

func (tl *tickerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width == 0 && size.Height == 0 {
		return
	}

	spacer := float32(-4)

	if tl.background.Size() != size {
		tl.background.Resize(size)
	}

	if tl.background.Position() != fyne.NewPos(0, 0) {
		tl.background.Move(fyne.NewPos(0, 0))
	}

	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{tl.title, tl.content, tl.status} {
		sz := obj.MinSize()
		if obj.Visible() && (sz.Height > 0 && sz.Width > 0) {
			centerItems = append(centerItems, obj)
		}
	}

	var totalHeight float32
	sizes := make([]fyne.Size, len(centerItems))
	for i, obj := range centerItems {
		s := obj.MinSize()
		sizes[i] = s
		totalHeight += s.Height
	}

	totalHeight += spacer * float32(len(centerItems)-1)

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	for i, obj := range centerItems {
		objSize := sizes[i]
		pos := fyne.NewPos((size.Width-objSize.Width)/2, currentY)

		if obj.Position() != pos {
			obj.Move(pos)
		}

		if obj.Size() != objSize {
			obj.Resize(objSize)
		}

		currentY += objSize.Height + spacer
	}
}

func (tl *tickerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if tickerDisplayLayoutCachedSize.Height == 0 {
		tickerDisplayLayoutCachedSize = fyne.NewSize(JC.UseTheme().Size(JC.SizeTickerWidth), JC.UseTheme().Size(JC.SizeTickerHeight))
	}
	return tickerDisplayLayoutCachedSize
}
