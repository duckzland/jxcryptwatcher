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

	list := JT.BT.Get()
	p := []*tickerDisplay{}

	for _, pot := range list {
		ticker := NewtickerDisplay(pot)
		ticker.Resize(fyne.NewSize(JC.TickerWidth, JC.TickerHeight))

		p = append(p, ticker)
	}

	tickers := make([]fyne.CanvasObject, len(p))
	for i := range p {
		tickers[i] = p[i]
	}

	grid := NewTickerContainer(
		&tickerGridLayout{
			minCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			dynCellSize:  fyne.NewSize(JC.TickerWidth, JC.TickerHeight),
			colCount:     1,
			rowCount:     1,
			innerPadding: JC.PanelPadding,
		},
		tickers,
	)

	JC.Tickers = container.NewStack(grid)

	JC.PrintMemUsage("End building tickers")

	return grid
}
