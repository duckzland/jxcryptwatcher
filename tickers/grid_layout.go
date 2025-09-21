package tickers

import (
	"math"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
)

type tickerGridLayout struct {
	minCellSize  fyne.Size
	dynCellSize  fyne.Size
	colCount     int
	rowCount     int
	innerPadding [4]float32 // top, right, bottom, left
	cWidth       float32
	minSize      fyne.Size
	dirty        bool
}

func (g *tickerGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if JA.LayoutManager == nil || JA.LayoutManager.ContainerSize().Width <= 0 || JA.LayoutManager.ContainerSize().Height <= 0 {
		return
	}

	if len(objects) == 0 {
		return
	}

	if size.Width == g.cWidth {
		return
	}

	g.cWidth = size.Width
	g.dirty = true

	hPad := g.innerPadding[1] + g.innerPadding[3] // right + left
	vPad := g.innerPadding[0] + g.innerPadding[2] // top + bottom

	g.colCount = 1
	g.rowCount = 0
	g.dynCellSize = g.minCellSize

	if g.minCellSize.Width > g.cWidth {
		g.minCellSize.Width = g.cWidth - hPad
	}

	if size.Width > g.minCellSize.Width {

		switch len(objects) {
		case 4:
			g.colCount = int(math.Floor(float64(size.Width+hPad) / float64(g.minCellSize.Width+hPad)))
			if g.colCount < 2 {
				g.colCount = 2
			} else if g.colCount > len(objects) {
				g.colCount = len(objects)
			}

			if g.colCount > 2 && g.colCount%2 != 0 {
				g.colCount--
			}

		default:
			g.colCount = len(objects)
		}

		pads := float32(0)
		for i := 0; i < g.colCount; i++ {
			pads += hPad

			if i == 0 {
				pads -= g.innerPadding[3]
			}

			if i == g.colCount-1 {
				pads -= g.innerPadding[1]
			}
		}

		emptySpace := size.Width - (float32(g.colCount) * g.minCellSize.Width) - pads
		if emptySpace > 0 {
			g.dynCellSize.Width += emptySpace / float32(g.colCount)
		}
	}

	if g.colCount == 0 {
		g.colCount = 1
	}

	if g.dynCellSize.Width > g.cWidth {
		g.dynCellSize.Width = g.cWidth
	}

	i, x, y := 0, g.innerPadding[3], g.innerPadding[0]

	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		if i%g.colCount == 0 {
			x = 0
			g.rowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.dynCellSize)

		if (i+1)%g.colCount == 0 {
			y += g.dynCellSize.Height + vPad
		}

		if (i+1)%g.colCount != 0 {
			x += g.dynCellSize.Width + hPad
		}

		i++
	}
}

func (g *tickerGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	if !g.dirty {
		return g.minSize
	}

	g.dirty = false

	aWidth := JA.LayoutManager.ContainerSize().Width

	c := int(math.Floor(float64(aWidth) / float64(g.dynCellSize.Width)))

	switch len(objects) {
	case 4:
		if c > 2 && c%2 != 0 {
			c--
		}
	default:
		c = len(objects)
	}

	r := int(math.Ceil(float64(len(objects)) / float64(c)))

	rows := max(r, 1)
	cols := max(c, 1)

	width := (g.dynCellSize.Width * float32(cols)) + (g.innerPadding[1] + g.innerPadding[3])
	height := (g.dynCellSize.Height * float32(rows)) + (float32(rows) * (g.innerPadding[0] + g.innerPadding[2]))

	g.minSize = fyne.NewSize(width, height)

	return g.minSize
}
