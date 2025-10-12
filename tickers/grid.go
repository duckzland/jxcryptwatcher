package tickers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var tickerGrid *tickerContainer = &tickerContainer{}

func RegisterTickerGrid() {
	JC.PrintMemUsage("Start building tickers")

	list := JT.UseTickerMaps().GetData()
	p := []*tickerDisplay{}

	for _, pot := range list {
		ticker := NewtickerDisplay(pot)
		ticker.Resize(fyne.NewSize(JC.UseTheme().Size(JC.SizeTickerWidth), JC.UseTheme().Size(JC.SizeTickerHeight)))

		p = append(p, ticker)
	}

	tickers := make([]fyne.CanvasObject, len(p))
	for i := range p {
		tickers[i] = p[i]
	}

	tickerGrid := NewTickerContainer(
		&tickerGridLayout{
			minCellSize: fyne.NewSize(JC.UseTheme().Size(JC.SizeTickerWidth), JC.UseTheme().Size(JC.SizeTickerHeight)),
			dynCellSize: fyne.NewSize(JC.UseTheme().Size(JC.SizeTickerWidth), JC.UseTheme().Size(JC.SizeTickerHeight)),
			colCount:    1,
			rowCount:    1,
			innerPadding: [4]float32{
				JC.UseTheme().Size(JC.SizePaddingTickerTop),
				JC.UseTheme().Size(JC.SizePaddingTickerRight),
				JC.UseTheme().Size(JC.SizePaddingTickerBottom),
				JC.UseTheme().Size(JC.SizePaddingTickerLeft),
			},
		},
		tickers,
	)

	JA.UseLayout().RegisterTickers(container.NewStack(tickerGrid))

	JC.PrintMemUsage("End building tickers")
}

func UseTickerGrid() *tickerContainer {
	return tickerGrid
}
