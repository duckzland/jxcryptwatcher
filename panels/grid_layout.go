package panels

import (
	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	"math"

	"fyne.io/fyne/v2"
)

type panelGridLayout struct {
	minCellSize  fyne.Size
	dynCellSize  fyne.Size
	colCount     int
	rowCount     int
	innerPadding [4]float32 // top, right, bottom, left
	objectCount  int
	cWidth       float32
	minSize      fyne.Size
	dirty        bool
}

func (g *panelGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if JA.LayoutManager == nil || JA.LayoutManager.ContainerSize().Width <= 0 || JA.LayoutManager.ContainerSize().Height <= 0 {
		return
	}

	if g.cWidth == size.Width && g.objectCount == len(objects) {
		return
	}

	g.cWidth = size.Width
	g.objectCount = len(objects)
	g.dirty = true

	hPad := g.innerPadding[1] + g.innerPadding[3] // right + left
	vPad := g.innerPadding[0] + g.innerPadding[2] // top + bottom

	g.colCount = 1
	g.rowCount = 0
	g.dynCellSize = g.minCellSize
	dragDropZones = []*panelDropZone{}

	sw := size.Width

	// Battling scrollbar, detect if we have scrollbar visible
	mr := g.countRows(size, hPad, objects)
	th := (g.dynCellSize.Height * float32(mr)) + (float32(mr) * (g.innerPadding[0] + g.innerPadding[2]))
	if th > JA.LayoutManager.Height() {
		sw -= 18
	}

	// Screen is too small for min width
	if g.minCellSize.Width > JA.LayoutManager.Width() {
		g.minCellSize.Width = JA.LayoutManager.Width() - hPad
	}

	if sw > g.minCellSize.Width {
		g.colCount = int(math.Floor(float64(sw+hPad) / float64(g.minCellSize.Width+hPad)))

		pads := float32(0)
		for i := 0; i < g.colCount; i++ {
			pads += hPad

			// Properly count pads, the first in column will not need left padding
			if i == 0 {
				pads -= g.innerPadding[3]
			}

			// Properly count pads, the last in column will not need right padding
			if i == g.colCount-1 {
				pads -= g.innerPadding[1]
			}
		}

		emptySpace := sw - (float32(g.colCount) * g.minCellSize.Width) - pads
		if emptySpace > 0 {
			g.dynCellSize.Width += emptySpace / float32(g.colCount)
		}
	}

	// Fix division by zero
	if g.colCount == 0 {
		g.colCount = 1
	}

	// Fix single column overflowing on android phone
	if g.dynCellSize.Width > JA.LayoutManager.Width() {
		g.dynCellSize.Width = JA.LayoutManager.Width()

		if th > JA.LayoutManager.Height() {
			g.dynCellSize.Width -= 18
		}
	}

	i, x, y := 0, g.innerPadding[3], g.innerPadding[0]

	if JA.DragPlaceholder != nil {
		if JA.DragPlaceholder.Size() != g.dynCellSize {
			JA.DragPlaceholder.Resize(g.dynCellSize)
		}
	}

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		// First in column, move to 0 horizontally
		if i%g.colCount == 0 {
			x = 0
			g.rowCount++
		}

		dz := panelDropZone{
			left:   x,
			right:  x + g.dynCellSize.Width,
			top:    y,
			bottom: y + g.dynCellSize.Height,
			panel:  child.(*panelDisplay),
		}

		dragDropZones = append(dragDropZones, &dz)

		pos := fyne.NewPos(x, y)

		if child.Position() != pos {
			child.Move(pos)
			JC.Logln("Grid layout moving child")
		}

		if child.Size() != g.dynCellSize {
			child.Resize(g.dynCellSize)
			JC.Logln("Grid layout resizing child")
		}

		// End of column, prepare to move down the next item
		if (i+1)%g.colCount == 0 {
			y += g.dynCellSize.Height + vPad
		}

		// Still in column, just move right horizontally
		if (i+1)%g.colCount != 0 {
			x += g.dynCellSize.Width + hPad
		}

		i++
	}
}

// Count approx how many rows will be, this isn't accurate and should be only used at the beginning of layouting
// After layouting use g.rowCount instead
func (g *panelGridLayout) countRows(size fyne.Size, hPad float32, objects []fyne.CanvasObject) int {

	r := 0
	i := 0
	c := int(math.Floor(float64(size.Width+hPad) / float64(g.minCellSize.Width+hPad)))

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if c != 0 && i%c == 0 {
			r++
		}

		i++
	}

	return r
}

func (g *panelGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	if !g.dirty {
		return g.minSize
	}

	rows := max(g.rowCount, 1)
	width := g.dynCellSize.Width
	height := (g.dynCellSize.Height * float32(rows)) + (float32(rows) * (g.innerPadding[0] + g.innerPadding[2]))

	// Battling scrollbar, when we have scrollbar give space for it
	if height > JC.MainLayoutContentHeight {
		width -= 16
	}

	g.minSize = fyne.NewSize(width, height)

	return g.minSize
}

func (g *panelGridLayout) Reset() {
	g.cWidth = 0
	g.objectCount = 0
}
