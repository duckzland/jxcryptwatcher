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

	grid := container.New(JL.NewDynamicGridWrapLayout(fyne.NewSize(300, 150)))
	list := JT.BP.Get()

	for i := range list {
		pkt := JT.BP.GetDataByIndex(i)
		pkt.Index = i
		grid.Add(createPanel(pkt))
	}

	JC.PrintMemUsage("End building panels")

	return grid
}
