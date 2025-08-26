package panels

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var DragPlaceholder fyne.CanvasObject

type PanelGridLayout struct {
	MinCellSize  fyne.Size
	DynCellSize  fyne.Size
	ColCount     int
	RowCount     int
	InnerPadding [4]float32 // top, right, bottom, left
}

func (g *PanelGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if JA.AppLayoutManager == nil || JA.AppLayoutManager.Width() <= 0 || JA.AppLayoutManager.Height() <= 0 {
		return
	}

	hPad := g.InnerPadding[1] + g.InnerPadding[3] // right + left
	vPad := g.InnerPadding[0] + g.InnerPadding[2] // top + bottom

	g.ColCount = 1
	g.RowCount = 0
	g.DynCellSize = g.MinCellSize

	// Battling scrollbar, detect if we have scrollbar visible
	mr := g.countRows(size, hPad, objects)
	th := (g.DynCellSize.Height * float32(mr)) + (float32(mr) * (g.InnerPadding[0] + g.InnerPadding[2]))
	if th > JA.AppLayoutManager.Height() {
		size.Width -= 18
	}

	// Screen is too small for min width
	if g.MinCellSize.Width > JA.AppLayoutManager.Width() {
		g.MinCellSize.Width = JA.AppLayoutManager.Width() - hPad
	}

	if size.Width > g.MinCellSize.Width {
		g.ColCount = int(math.Floor(float64(size.Width+hPad) / float64(g.MinCellSize.Width+hPad)))

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

	i, x, y := 0, g.InnerPadding[3], g.InnerPadding[0]

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if child == DragPlaceholder && DragPlaceholder != nil {
			child.Resize(g.DynCellSize)
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

		if child == DragPlaceholder && DragPlaceholder != nil {
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

	rows := max(g.RowCount, 1)
	width := g.DynCellSize.Width
	height := (g.DynCellSize.Height * float32(rows)) + (float32(rows) * (g.InnerPadding[0] + g.InnerPadding[2]))

	// Battling scrollbar, when we have scrollbar give space for it
	if height > JC.MainLayoutContentHeight {
		width -= 16
	}

	return fyne.NewSize(width, height)
}

type CreatePanelFunc func(*JT.PanelDataType) fyne.CanvasObject

func NewPanelGridLayout(size fyne.Size, padding [4]float32) fyne.Layout {
	return &PanelGridLayout{
		MinCellSize:  size,
		DynCellSize:  size,
		ColCount:     1,
		RowCount:     1,
		InnerPadding: padding,
	}
}

func NewPanelGrid(createPanel CreatePanelFunc) *fyne.Container {
	JC.PrintMemUsage("Start building panels")

	// Get the list of panel data
	list := JT.BP.Get()
	p := []*PanelDisplay{}

	for _, pot := range list {
		// Retrieve and initialize panel data
		pkt := JT.BP.GetData(pot.ID)
		pkt.Status = 0

		// Create the panel
		panel := createPanel(pkt).(*PanelDisplay)
		panel.Resize(fyne.NewSize(JC.PanelWidth, JC.PanelHeight))

		p = append(p, panel)
	}

	o := make([]fyne.CanvasObject, len(p))
	for i := range p {
		o[i] = p[i]
	}

	// Using direct spread injection for objects to save multiple refresh calls
	grid := container.New(NewPanelGridLayout(
		fyne.NewSize(JC.PanelWidth, JC.PanelHeight),
		JC.PanelPadding,
	), o...)

	// Global dummy panel for placeholder
	DragPlaceholder = canvas.NewRectangle(JC.PanelPlaceholderBG)
	DragPlaceholder.Hide()

	if rect, ok := DragPlaceholder.(*canvas.Rectangle); ok {
		rect.CornerRadius = JC.PanelBorderRadius
	}

	JC.PrintMemUsage("End building panels")

	return grid
}
