package panels

import (
	"math"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
)

type panelGridLayout struct {
	minCellSize  fyne.Size
	dynCellSize  fyne.Size
	colCount     int
	rowCount     int
	innerPadding [4]float32 // top, right, bottom, left
	objectCount  int
	objects      []fyne.CanvasObject
	cWidth       float32
	cHeight      float32
	minSize      fyne.Size
	dirty        bool
}

func (g *panelGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if JA.UseLayout() == nil || JA.UseLayout().UseContainer().Size().Width <= 0 || JA.UseLayout().UseContainer().Size().Height <= 0 {
		return
	}

	if g.cWidth == size.Width && g.cHeight == size.Height && g.objectCount == len(objects) {
		return
	}

	g.cWidth = size.Width
	g.cHeight = size.Height
	g.objectCount = len(objects)
	g.objects = objects
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
	if th > JA.UseLayout().UseScroll().Size().Height {
		sw -= 18
	}

	// Screen is too small for min width
	if g.minCellSize.Width > JA.UseLayout().UseScroll().Size().Width {
		g.minCellSize.Width = JA.UseLayout().UseScroll().Size().Width - hPad
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
	if g.dynCellSize.Width > JA.UseLayout().UseScroll().Size().Width {
		g.dynCellSize.Width = JA.UseLayout().UseScroll().Size().Width

		if th > JA.UseLayout().UseScroll().Size().Height {
			g.dynCellSize.Width -= 18
		}
	}

	i, x, y := 0, g.innerPadding[3], float32(0)

	JA.UseLayout().UsePlaceholder().Resize(g.dynCellSize)

	for _, child := range objects {

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
		}

		if child.Size() != g.dynCellSize {
			child.Resize(g.dynCellSize)
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

	g.OnScrolled(JA.UseLayout().UseScroll().Offset)
}

// Count approx how many rows will be, this isn't accurate and should be only used at the beginning of layouting
// After layouting use g.rowCount instead
func (g *panelGridLayout) countRows(size fyne.Size, hPad float32, objects []fyne.CanvasObject) int {

	r := 0
	c := int(math.Floor(float64(size.Width+hPad) / float64(g.minCellSize.Width+hPad)))

	for i := range objects {
		if c != 0 && i%c == 0 {
			r++
		}
	}

	return r
}

func (g *panelGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if !g.dirty {
		return g.minSize
	}

	g.dirty = false

	rows := max(g.rowCount, 1)
	width := g.dynCellSize.Width
	height := (g.dynCellSize.Height * float32(rows)) + (float32(rows) * (g.innerPadding[0] + g.innerPadding[2]))
	height -= g.innerPadding[0] + g.innerPadding[2]

	// Battling scrollbar, when we have scrollbar give space for it
	if height > JA.UseLayout().UseScroll().Size().Height {
		width -= 18
	}

	g.minSize = fyne.NewSize(width, height)

	return g.minSize
}

func (g *panelGridLayout) Reset() {
	g.cWidth = 0
	g.objectCount = 0
}

func (g *panelGridLayout) OnScrolled(pos fyne.Position) {

	sHeight := JA.UseLayout().UseScroll().Size().Height
	vPad := g.innerPadding[0] + g.innerPadding[2]
	miny := pos.Y - g.dynCellSize.Height
	maxy := sHeight + pos.Y + vPad

	for _, child := range g.objects {
		y := child.Position().Y
		if y > maxy || miny > y {
			if child.Visible() {
				child.Hide()
			}
		} else {
			if !child.Visible() {
				child.Show()
			}
		}
	}
}
