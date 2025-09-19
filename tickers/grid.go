package tickers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var Grid *tickerContainer = &tickerContainer{}

func NewTickerGrid() *tickerContainer {
	JC.PrintMemUsage("Start building tickers")

	// Get the list of panel data
	list := JT.BT.Get()
	p := []*tickerDisplay{}

	for _, pot := range list {
		// Create the panel
		ticker := NewtickerDisplay(pot)
		ticker.Resize(fyne.NewSize(JC.TickerWidth, JC.TickerHeight))

		p = append(p, ticker)
	}

	o := make([]fyne.CanvasObject, len(p))
	for i := range p {
		o[i] = p[i]
	}

	// Using direct spread injection for objects to save multiple refresh calls
	grid := NewTickerContainer(
		&tickerGridLayout{
			MinCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			DynCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			ColCount:     1,
			RowCount:     1,
			InnerPadding: JC.PanelPadding,
		},
		o,
	)

	JC.Tickers = container.NewStack(grid)
	JC.PrintMemUsage("End building tickers")

	return grid
}
