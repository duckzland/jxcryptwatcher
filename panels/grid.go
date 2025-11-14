package panels

import (
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
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
	uuid   string
}

func RegisterPanelGrid(createPanel CreatePanelFunc) {
	JC.PrintPerfStats("Generating Panels", time.Now())

	// Get the list of panel data
	list := JT.UsePanelMaps().GetData()
	p := []*panelDisplay{}

	for _, pot := range list {
		// Retrieve and initialize panel data
		pkt := JT.UsePanelMaps().GetDataByID(pot.GetID())

		if pkt.UsePanelKey().IsValueMatchingFloat(-1, JC.STRING_DOUBLE_EQUAL) {
			pkt.SetStatus(JC.STATE_LOADING)
		}

		// Create the panel
		panel := createPanel(pkt).(*panelDisplay)
		panel.Resize(fyne.NewSize(JC.UseTheme().Size(JC.SizePanelWidth), JC.UseTheme().Size(JC.SizePanelHeight)))

		p = append(p, panel)
	}

	panels := make([]fyne.CanvasObject, len(p))
	for i := range p {
		panels[i] = p[i]
	}

	layout := &panelGridLayout{
		minCellSize: fyne.NewSize(JC.UseTheme().Size(JC.SizePanelWidth), JC.UseTheme().Size(JC.SizePanelHeight)),
		dynCellSize: fyne.NewSize(JC.UseTheme().Size(JC.SizePanelWidth), JC.UseTheme().Size(JC.SizePanelHeight)),
		colCount:    1,
		rowCount:    1,
		innerPadding: [4]float32{
			JC.UseTheme().Size(JC.SizePaddingPanelTop),
			JC.UseTheme().Size(JC.SizePaddingPanelRight),
			JC.UseTheme().Size(JC.SizePaddingPanelBottom),
			JC.UseTheme().Size(JC.SizePaddingPanelLeft),
		},
	}

	panelGrid = NewPanelContainer(
		layout,
		panels,
	)

	// Global dummy panel for placeholder
	dragDropZones = []*panelDropZone{}

	JA.UseLayout().UseScroll().OnScrolled = layout.OnScrolled
}

func UsePanelGrid() *panelContainer {
	return panelGrid
}
