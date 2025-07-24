package layouts

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Layout = (*dynamicGridWrapLayout)(nil)

type dynamicGridWrapLayout struct {
	MinCellSize fyne.Size
	DynCellSize fyne.Size
	colCount    int
	rowCount    int
}

func NewDynamicGridWrapLayout(size fyne.Size) fyne.Layout {
	return &dynamicGridWrapLayout{size, size, 1, 1}
}

func (g *dynamicGridWrapLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	padding := theme.Padding()
	g.colCount = 1
	g.rowCount = 0

	// Reset first
	g.DynCellSize = g.MinCellSize

	if size.Width > g.MinCellSize.Width {
		g.colCount = int(math.Floor(float64(size.Width+padding) / float64(g.MinCellSize.Width+padding)))

		// Grap empty space and spread them to all cells
		emptySpace := size.Width - (float32(g.colCount) * g.MinCellSize.Width) - (float32(padding) * float32(g.colCount))
		if emptySpace > float32(0) {
			g.DynCellSize.Width += emptySpace / float32(g.colCount)
		}
	}

	i, x, y := 0, float32(0), float32(0)
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
			x = 0
			y += g.DynCellSize.Height + padding
		} else {
			x += g.DynCellSize.Width + padding
		}
		i++
	}
}

func (g *dynamicGridWrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return fyne.NewSize(g.DynCellSize.Width,
		(g.DynCellSize.Height*float32(rows))+(float32(rows-1)*theme.Padding()))
}
