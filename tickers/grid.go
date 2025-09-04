package tickers

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type TickerGridLayout struct {
	MinCellSize  fyne.Size
	DynCellSize  fyne.Size
	ColCount     int
	RowCount     int
	InnerPadding [4]float32 // top, right, bottom, left
}

func (g *TickerGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if JA.AppLayoutManager == nil || JA.AppLayoutManager.Width() <= 0 || JA.AppLayoutManager.Height() <= 0 {
		return
	}

	hPad := g.InnerPadding[1] + g.InnerPadding[3] // right + left
	vPad := g.InnerPadding[0] + g.InnerPadding[2] // top + bottom

	g.ColCount = 1
	g.RowCount = 0
	g.DynCellSize = g.MinCellSize

	// Screen is too small for min width
	if g.MinCellSize.Width > JA.AppLayoutManager.Width() {
		g.MinCellSize.Width = JA.AppLayoutManager.Width() - hPad
	}

	if size.Width > g.MinCellSize.Width {
		g.ColCount = int(math.Floor(float64(size.Width+hPad) / float64(g.MinCellSize.Width+hPad)))

		if len(objects) < g.ColCount {
			g.ColCount = len(objects)
		}

		pads := float32(0)
		for i := 0; i < g.ColCount; i++ {
			pads += hPad

			// Properly count pads, the first in column will not need left padding
			if i == 0 {
				pads -= g.InnerPadding[3]
			}

			// Properly count pads, the last in column will not need right padding
			if i == g.ColCount-1 {
				pads -= g.InnerPadding[1]
			}
		}

		emptySpace := size.Width - (float32(g.ColCount) * g.MinCellSize.Width) - pads
		if emptySpace > 0 {
			g.DynCellSize.Width += emptySpace / float32(g.ColCount)
		}
	}

	// Fix division by zero
	if g.ColCount == 0 {
		g.ColCount = 1
	}

	// Fix single column overflowing on android phone
	if g.DynCellSize.Width > JA.AppLayoutManager.Width() {
		g.DynCellSize.Width = JA.AppLayoutManager.Width()
	}

	i, x, y := 0, g.InnerPadding[3], g.InnerPadding[0]

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		// First in column, move to 0 horizontally
		if i%g.ColCount == 0 {
			x = 0
			g.RowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.DynCellSize)

		// End of column, prepare to move down the next item
		if (i+1)%g.ColCount == 0 {
			y += g.DynCellSize.Height + vPad
		}

		// Still in column, just move right horizontally
		if (i+1)%g.ColCount != 0 {
			x += g.DynCellSize.Width + hPad
		}

		i++
	}
}

func (g *TickerGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	rows := max(g.RowCount, 1)
	width := g.DynCellSize.Width
	height := (g.DynCellSize.Height * float32(rows)) + (float32(rows) * (g.InnerPadding[0] + g.InnerPadding[2]))

	return fyne.NewSize(width, height)
}

func NewTickerGrid() *fyne.Container {
	JC.PrintMemUsage("Start building tickers")

	if !JT.Config.HasProKey() || !JA.AppStatusManager.IsValidProKey() {
		JC.Logln("Refusing to create tickers due to no pro key")
		return nil
	}

	// Get the list of panel data
	list := JT.BT.Get()
	p := []*TickerDisplay{}

	for _, pot := range list {
		// Create the panel
		ticker := NewTickerDisplay(pot)
		ticker.Resize(fyne.NewSize(JC.TickerWidth, JC.TickerHeight))

		p = append(p, ticker)
	}

	o := make([]fyne.CanvasObject, len(p))
	for i := range p {
		o[i] = p[i]
	}

	// Using direct spread injection for objects to save multiple refresh calls
	grid := container.New(
		&TickerGridLayout{
			MinCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			DynCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			ColCount:     1,
			RowCount:     1,
			InnerPadding: JC.PanelPadding,
		},
		o...,
	)

	JC.PrintMemUsage("End building tickers")

	return grid
}
