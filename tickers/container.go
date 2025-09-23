package tickers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JT "jxwatcher/types"
)

type tickerContainer struct {
	widget.BaseWidget
	Objects []fyne.CanvasObject
	layout  *tickerGridLayout
}

func (c *tickerContainer) Add(obj fyne.CanvasObject) {
	c.Objects = append(c.Objects, obj)
}

func (c *tickerContainer) Remove(obj fyne.CanvasObject) {
	for i, o := range c.Objects {
		if o == obj {
			c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
			break
		}
	}
}

func (c *tickerContainer) CreateRenderer() fyne.WidgetRenderer {
	return &tickerContainerLayout{
		container: c,
	}
}

func (c *tickerContainer) UpdateTickersContent(shouldUpdate func(pdt JT.TickerData) bool) {
	for _, obj := range c.Objects {
		if ticker, ok := obj.(*tickerDisplay); ok {

			pdt := JT.UseTickerMaps().GetData(ticker.GetTag())

			if shouldUpdate != nil && !shouldUpdate(pdt) {
				continue
			}

			ticker.updateContent()
		}
	}
}

func NewTickerContainer(
	layout *tickerGridLayout,
	Objects []fyne.CanvasObject,
) *tickerContainer {
	c := &tickerContainer{
		Objects: Objects,
		layout:  layout,
	}

	c.ExtendBaseWidget(c)

	return c
}
