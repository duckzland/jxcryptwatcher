package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var panelDisplayLayoutCachedSize fyne.Size

type panelDisplayLayout struct {
	background *canvas.Rectangle
	title      *panelText
	content    *panelText
	subtitle   *panelText
	bottomText *panelText
	action     *panelAction
}

func (pl *panelDisplayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width <= 0 && size.Height <= 0 {
		return
	}

	spacer := float32(-2)

	if pl.background.Size() != size {
		pl.background.Resize(size)
	}

	if pl.background.Position() != fyne.NewPos(0, 0) {
		pl.background.Move(fyne.NewPos(0, 0))
	}

	centerItems := []fyne.CanvasObject{}
	sizes := []fyne.Size{}
	totalHeight := float32(0)

	for _, obj := range []fyne.CanvasObject{pl.title, pl.content, pl.subtitle, pl.bottomText} {
		if obj.Visible() {
			sz := obj.MinSize()
			if sz.Width > 0 && sz.Height > 0 {
				centerItems = append(centerItems, obj)
				sizes = append(sizes, sz)
				totalHeight += sz.Height
			}
		}
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

	actionSize := pl.action.MinSize()
	actionPos := fyne.NewPos(size.Width-actionSize.Width, 0)

	if pl.action.Position() != actionPos {
		pl.action.Move(actionPos)
	}

	if pl.action.Size() != actionSize {
		pl.action.Resize(actionSize)
	}
}

func (pl *panelDisplayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if panelDisplayLayoutCachedSize.Height == 0 {
		panelDisplayLayoutCachedSize = fyne.NewSize(JC.UseTheme().Size(JC.SizePanelWidth), JC.UseTheme().Size(JC.SizePanelHeight))
	}

	return panelDisplayLayoutCachedSize
}
