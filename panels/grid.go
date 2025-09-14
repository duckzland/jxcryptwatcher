package panels

import (
	"math"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var DragDropZones []*PanelDropZone
var Grid *PanelGridContainer = &PanelGridContainer{}

type PanelDropZone struct {
	top    float32
	left   float32
	bottom float32
	right  float32
	panel  *PanelDisplay
}

type PanelGridLayout struct {
	MinCellSize  fyne.Size
	DynCellSize  fyne.Size
	ColCount     int
	RowCount     int
	InnerPadding [4]float32 // top, right, bottom, left
	objectCount  int
	cWidth       float32
	minSize      fyne.Size
	dirty        bool
}

func (g *PanelGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if JA.AppLayoutManager == nil || JA.AppLayoutManager.ContainerSize().Width <= 0 || JA.AppLayoutManager.ContainerSize().Height <= 0 {
		return
	}

	if g.cWidth == size.Width && g.objectCount == len(objects) {
		return
	}

	g.cWidth = size.Width
	g.objectCount = len(objects)
	g.dirty = true

	hPad := g.InnerPadding[1] + g.InnerPadding[3] // right + left
	vPad := g.InnerPadding[0] + g.InnerPadding[2] // top + bottom

	g.ColCount = 1
	g.RowCount = 0
	g.DynCellSize = g.MinCellSize
	DragDropZones = []*PanelDropZone{}

	sw := size.Width

	// Battling scrollbar, detect if we have scrollbar visible
	mr := g.countRows(size, hPad, objects)
	th := (g.DynCellSize.Height * float32(mr)) + (float32(mr) * (g.InnerPadding[0] + g.InnerPadding[2]))
	if th > JA.AppLayoutManager.Height() {
		sw -= 18
	}

	// Screen is too small for min width
	if g.MinCellSize.Width > JA.AppLayoutManager.Width() {
		g.MinCellSize.Width = JA.AppLayoutManager.Width() - hPad
	}

	if sw > g.MinCellSize.Width {
		g.ColCount = int(math.Floor(float64(sw+hPad) / float64(g.MinCellSize.Width+hPad)))

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

		emptySpace := sw - (float32(g.ColCount) * g.MinCellSize.Width) - pads
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

		if th > JA.AppLayoutManager.Height() {
			g.DynCellSize.Width -= 18
		}
	}

	i, x, y := 0, g.InnerPadding[3], g.InnerPadding[0]

	if JA.DragPlaceholder != nil {
		JA.DragPlaceholder.Resize(g.DynCellSize)
	}

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		// First in column, move to 0 horizontally
		if i%g.ColCount == 0 {
			x = 0
			g.RowCount++
		}

		dz := PanelDropZone{
			left:   x,
			right:  x + g.DynCellSize.Width,
			top:    y,
			bottom: y + g.DynCellSize.Height,
			panel:  child.(*PanelDisplay),
		}

		DragDropZones = append(DragDropZones, &dz)

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

// Count approx how many rows will be, this isn't accurate and should be only used at the beginning of layouting
// After layouting use g.RowCount instead
func (g *PanelGridLayout) countRows(size fyne.Size, hPad float32, objects []fyne.CanvasObject) int {

	r := 0
	i := 0
	c := int(math.Floor(float64(size.Width+hPad) / float64(g.MinCellSize.Width+hPad)))

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

func (g *PanelGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	if !g.dirty {
		return g.minSize
	}

	rows := max(g.RowCount, 1)
	width := g.DynCellSize.Width
	height := (g.DynCellSize.Height * float32(rows)) + (float32(rows) * (g.InnerPadding[0] + g.InnerPadding[2]))

	// Battling scrollbar, when we have scrollbar give space for it
	if height > JC.MainLayoutContentHeight {
		width -= 16
	}

	g.minSize = fyne.NewSize(width, height)

	return g.minSize
}

func (g *PanelGridLayout) Reset() {
	g.cWidth = 0
	g.objectCount = 0
}

type CreatePanelFunc func(*JT.PanelDataType) fyne.CanvasObject

func NewPanelGrid(createPanel CreatePanelFunc) *PanelGridContainer {
	JC.PrintMemUsage("Start building panels")

	// Get the list of panel data
	list := JT.BP.GetData()
	p := []*PanelDisplay{}

	for _, pot := range list {
		// Retrieve and initialize panel data
		pkt := JT.BP.GetDataByID(pot.GetID())

		if pkt.UsePanelKey().GetValueFloat() == -1 {
			pkt.SetStatus(JC.STATE_LOADING)
		}

		// Create the panel
		panel := createPanel(pkt).(*PanelDisplay)
		panel.Resize(fyne.NewSize(JC.PanelWidth, JC.PanelHeight))

		p = append(p, panel)
	}

	o := make([]fyne.CanvasObject, len(p))
	for i := range p {
		o[i] = p[i]
	}

	grid := NewPanelGridContainer(
		&PanelGridLayout{
			MinCellSize:  fyne.NewSize(JC.PanelWidth, JC.PanelHeight),
			DynCellSize:  fyne.NewSize(JC.PanelWidth, JC.PanelHeight),
			ColCount:     1,
			RowCount:     1,
			InnerPadding: JC.PanelPadding,
		},
		o,
	)

	// Global dummy panel for placeholder
	DragDropZones = []*PanelDropZone{}

	JC.PrintMemUsage("End building panels")

	return grid
}
