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
		ticker.Resize(fyne.NewSize(JC.ThemeSize(JC.SizeTickerWidth), JC.ThemeSize(JC.SizeTickerHeight)))

		p = append(p, ticker)
	}

	tickers := make([]fyne.CanvasObject, len(p))
	for i := range p {
		tickers[i] = p[i]
	}

	grid := NewTickerContainer(
		&tickerGridLayout{
			minCellSize: fyne.NewSize(JC.ThemeSize(JC.SizeTickerWidth), JC.ThemeSize(JC.SizeTickerHeight)),
			dynCellSize: fyne.NewSize(JC.ThemeSize(JC.SizeTickerWidth), JC.ThemeSize(JC.SizeTickerHeight)),
			colCount:    1,
			rowCount:    1,
			innerPadding: [4]float32{
				JC.ThemeSize(JC.SizePaddingPanelTop),
				JC.ThemeSize(JC.SizePaddingPanelRight),
				JC.ThemeSize(JC.SizePaddingPanelBottom),
				JC.ThemeSize(JC.SizePaddingPanelLeft),
			},
		},
		tickers,
	)

	JC.Tickers = container.NewStack(grid)

	JC.PrintMemUsage("End building tickers")

	return grid
}
