package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JL "jxwatcher/layouts"
	JT "jxwatcher/types"
)

type CreatePanelFunc func(*JT.PanelDataType) fyne.CanvasObject

func NewPanelGrid(createPanel CreatePanelFunc) *fyne.Container {

	JC.PrintMemUsage("Start building panels")

	grid := container.New(JL.NewDynamicGridWrapLayout(fyne.NewSize(JC.PanelWidth, JC.PanelHeight), JC.PanelPadding))
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
