package tickers

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type tickerGridRenderer struct {
	container *TickerGridContainer
}

func (r *tickerGridRenderer) Layout(size fyne.Size) {
	r.container.layout.Layout(r.container.Objects, size)
}

func (r *tickerGridRenderer) MinSize() fyne.Size {
	return r.container.layout.MinSize(r.container.Objects)
}

func (r *tickerGridRenderer) Refresh() {
	JC.MainDebouncer.Call("ticker_container_refresh", 10*time.Millisecond, func() {
		fyne.Do(func() {
			r.Layout(r.container.Size())
		})
	})
}

func (r *tickerGridRenderer) Objects() []fyne.CanvasObject {
	return r.container.Objects
}

func (r *tickerGridRenderer) Destroy() {
}

type TickerGridContainer struct {
	widget.BaseWidget
	Objects []fyne.CanvasObject
	layout  *TickerGridLayout
}

func (c *TickerGridContainer) Add(obj fyne.CanvasObject) {
	c.Objects = append(c.Objects, obj)
}

func (c *TickerGridContainer) Remove(obj fyne.CanvasObject) {
	for i, o := range c.Objects {
		if o == obj {
			c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
			break
		}
	}
}

func (c *TickerGridContainer) CreateRenderer() fyne.WidgetRenderer {
	return &tickerGridRenderer{
		container: c,
	}
}

func (c *TickerGridContainer) UpdateTickersContent(shouldUpdate func(pdt *JT.TickerDataType) bool) {
	for _, obj := range c.Objects {
		if ticker, ok := obj.(*TickerDisplay); ok {

			pdt := JT.BT.GetData(ticker.GetTag())

			if shouldUpdate != nil && !shouldUpdate(pdt) {
				continue
			}

			ticker.UpdateContent()
		}
	}
}
func NewTickerGridContainer(
	layout *TickerGridLayout,
	Objects []fyne.CanvasObject,
) *TickerGridContainer {
	c := &TickerGridContainer{
		Objects: Objects,
		layout:  layout,
	}

	c.ExtendBaseWidget(c)

	return c
}
