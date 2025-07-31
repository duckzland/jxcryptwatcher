package panels

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type PanelGridLayout struct {
	MinCellSize  fyne.Size
	DynCellSize  fyne.Size
	colCount     int
	rowCount     int
	InnerPadding [4]float32 // top, right, bottom, left
}

func (g *PanelGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	hPad := g.InnerPadding[1] + g.InnerPadding[3] // right + left
	vPad := g.InnerPadding[0] + g.InnerPadding[2] // top + bottom

	g.colCount = 1
	g.rowCount = 0
	g.DynCellSize = g.MinCellSize

	// Mobile or super small screen
	if size.Width != 0 && int64(size.Width) < int64(300) {
		g.MinCellSize.Width = 200
	}

	if size.Width > g.MinCellSize.Width {
		g.colCount = int(math.Floor(float64(size.Width+hPad) / float64(g.MinCellSize.Width+hPad)))

		pads := float32(0)
		for i := 0; i < g.colCount; i++ {
			pads += hPad

			// Properly count pads, the first in column will not need left padding
			if i == 0 {
				pads -= g.InnerPadding[3]
			}

			// Properly count pads, the last in column will not need right padding
			if i == g.colCount-1 {
				pads -= g.InnerPadding[1]
			}
		}

		emptySpace := size.Width - (float32(g.colCount) * g.MinCellSize.Width) - pads
		if emptySpace > 0 {
			g.DynCellSize.Width += emptySpace / float32(g.colCount)
		}
	}

	i, x, y := 0, g.InnerPadding[3], g.InnerPadding[0]

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		// First in column, move to 0 horizontally
		if i%g.colCount == 0 {
			x = 0
			g.rowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.DynCellSize)

		// End of column, prepare to move down the next item
		if (i+1)%g.colCount == 0 {
			y += g.DynCellSize.Height + vPad
		}

		// Still in column, just move right horizontally
		if (i+1)%g.colCount != 0 {
			x += g.DynCellSize.Width + hPad
		}

		i++
	}
}

func (g *PanelGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := max(g.rowCount, 1)

	return fyne.NewSize(
		g.DynCellSize.Width,
		(g.DynCellSize.Height*float32(rows))+(float32(rows-1)*g.InnerPadding[0])+g.InnerPadding[2],
	)
}

type CreatePanelFunc func(*JT.PanelDataType) fyne.CanvasObject

func NewPanelGridLayout(size fyne.Size, padding [4]float32) fyne.Layout {
	return &PanelGridLayout{
		MinCellSize:  size,
		DynCellSize:  size,
		colCount:     1,
		rowCount:     1,
		InnerPadding: padding,
	}
}

func NewPanelGrid(createPanel CreatePanelFunc) *fyne.Container {

	JC.PrintMemUsage("Start building panels")

	grid := container.New(NewPanelGridLayout(fyne.NewSize(JC.PanelWidth, JC.PanelHeight), JC.PanelPadding))
	list := JT.BP.Get()

	for i := range list {
		// Using index for first initial boot
		pkt := JT.BP.GetDataByIndex(i)
		pkt.Status = 0
		grid.Add(createPanel(pkt))
	}

	JC.PrintMemUsage("End building panels")

	return grid
}
