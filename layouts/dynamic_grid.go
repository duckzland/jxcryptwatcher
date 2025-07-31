package layouts

import (
	"math"

	"fyne.io/fyne/v2"
)

var _ fyne.Layout = (*dynamicGridWrapLayout)(nil)

type dynamicGridWrapLayout struct {
	MinCellSize  fyne.Size
	DynCellSize  fyne.Size
	colCount     int
	rowCount     int
	InnerPadding [4]float32 // top, right, bottom, left
}

func NewDynamicGridWrapLayout(size fyne.Size, padding [4]float32) fyne.Layout {
	return &dynamicGridWrapLayout{
		MinCellSize:  size,
		DynCellSize:  size,
		colCount:     1,
		rowCount:     1,
		InnerPadding: padding,
	}
}

func (g *dynamicGridWrapLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	hPad := g.InnerPadding[1] + g.InnerPadding[3] // right + left
	vPad := g.InnerPadding[0] + g.InnerPadding[2] // top + bottom

	g.colCount = 1
	g.rowCount = 0
	g.DynCellSize = g.MinCellSize

	if size.Width > g.MinCellSize.Width {
		g.colCount = int(math.Floor(float64(size.Width+hPad) / float64(g.MinCellSize.Width+hPad)))
		emptySpace := size.Width - (float32(g.colCount) * g.MinCellSize.Width) - (float32(g.colCount) * hPad)
		if emptySpace > 0 {
			g.DynCellSize.Width += emptySpace / float32(g.colCount)
		}
	}

	i, x, y := 0, g.InnerPadding[3], g.InnerPadding[0]
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if i%g.colCount == 0 {
			g.rowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.DynCellSize)

		if (i+1)%g.colCount == 0 {
			x = g.InnerPadding[3]
			y += g.DynCellSize.Height + vPad
		} else {
			x += g.DynCellSize.Width + hPad
		}
		i++
	}
}

func (g *dynamicGridWrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return fyne.NewSize(
		g.DynCellSize.Width,
		(g.DynCellSize.Height*float32(rows))+(float32(rows-1)*g.InnerPadding[0])+g.InnerPadding[2],
	)
}
