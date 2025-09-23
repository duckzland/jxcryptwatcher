package panels

import (
	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var dragDropZones []*panelDropZone
var panelGrid *panelContainer = &panelContainer{}

type CreatePanelFunc func(JT.PanelData) fyne.CanvasObject

type panelDropZone struct {
	top    float32
	left   float32
	bottom float32
	right  float32
	panel  *panelDisplay
}

func RegisterPanelGrid(createPanel CreatePanelFunc) {
	JC.PrintMemUsage("Start building panels")

	// Get the list of panel data
	list := JT.UsePanelMaps().GetData()
	p := []*panelDisplay{}

	for _, pot := range list {
		// Retrieve and initialize panel data
		pkt := JT.UsePanelMaps().GetDataByID(pot.GetID())

		if pkt.UsePanelKey().GetValueFloat() == -1 {
			pkt.SetStatus(JC.STATE_LOADING)
		}

		// Create the panel
		panel := createPanel(pkt).(*panelDisplay)
		panel.Resize(fyne.NewSize(JC.ThemeSize(JC.SizePanelWidth), JC.ThemeSize(JC.SizePanelHeight)))

		p = append(p, panel)
	}

	panels := make([]fyne.CanvasObject, len(p))
	for i := range p {
		panels[i] = p[i]
	}

	panelGrid = NewpanelContainer(
		&panelGridLayout{
			minCellSize: fyne.NewSize(JC.ThemeSize(JC.SizePanelWidth), JC.ThemeSize(JC.SizePanelHeight)),
			dynCellSize: fyne.NewSize(JC.ThemeSize(JC.SizePanelWidth), JC.ThemeSize(JC.SizePanelHeight)),
			colCount:    1,
			rowCount:    1,
			innerPadding: [4]float32{
				JC.ThemeSize(JC.SizePaddingPanelTop),
				JC.ThemeSize(JC.SizePaddingPanelRight),
				JC.ThemeSize(JC.SizePaddingPanelBottom),
				JC.ThemeSize(JC.SizePaddingPanelLeft),
			},
		},
		panels,
	)

	// Global dummy panel for placeholder
	dragDropZones = []*panelDropZone{}

	JC.PrintMemUsage("End building panels")
}

func UsePanelGrid() *panelContainer {
	return panelGrid
}
