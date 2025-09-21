package panels

import (
	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var dragDropZones []*panelDropZone
var Grid *panelContainer = &panelContainer{}

type CreatePanelFunc func(*JT.PanelDataType) fyne.CanvasObject

type panelDropZone struct {
	top    float32
	left   float32
	bottom float32
	right  float32
	panel  *PanelDisplay
}

func NewPanelGrid(createPanel CreatePanelFunc) *panelContainer {
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

	panels := make([]fyne.CanvasObject, len(p))
	for i := range p {
		panels[i] = p[i]
	}

	grid := NewpanelContainer(
		&panelGridLayout{
			minCellSize:  fyne.NewSize(JC.PanelWidth, JC.PanelHeight),
			dynCellSize:  fyne.NewSize(JC.PanelWidth, JC.PanelHeight),
			colCount:     1,
			rowCount:     1,
			innerPadding: JC.PanelPadding,
		},
		panels,
	)

	// Global dummy panel for placeholder
	dragDropZones = []*panelDropZone{}

	JC.PrintMemUsage("End building panels")

	return grid
}
