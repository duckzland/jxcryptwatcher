package layouts

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*dynamicGridWrapLayout)(nil)

type dynamicGridWrapLayout struct {
	MinCellSize fyne.Size
	DynCellSize fyne.Size
	colCount    int
	rowCount    int
}

// NewdynamicGridWrapLayout returns a new dynamicGridWrapLayout instance
func NewDynamicGridWrapLayout(size fyne.Size) fyne.Layout {
	return &dynamicGridWrapLayout{size, size, 1, 1}
}

// Layout is called to pack all child objects into a specified size.
// For a dynamicGridWrapLayout this will attempt to lay all the child objects in a row
// and wrap to a new row if the size is not large enough.
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

// MinSize finds the smallest size that satisfies all the child objects.
// For a dynamicGridWrapLayout this is simply the specified MinCellSize as a single column
// layout has no padding. The returned size does not take into account the number
// of columns as this layout re-flows dynamically.
func (g *dynamicGridWrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return fyne.NewSize(g.DynCellSize.Width,
		(g.DynCellSize.Height*float32(rows))+(float32(rows-1)*theme.Padding()))
}
