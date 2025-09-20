package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
)

type AppMainLayout struct {
	padding     float32
	topBar      *fyne.Container
	tickers     *fyne.Container
	overlay     *fyne.Container
	content     *container.Scroll
	placeholder *canvas.Rectangle
}

func (a *AppMainLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	if size.Width <= 0 || size.Height <= 0 || a.topBar == nil || a.content == nil {
		return
	}

	const splitThreshold = 1600.0
	const minTickerWidth = 700.0

	padding := a.padding

	var topHeight fyne.Size
	var tickerHeight fyne.Size
	var contentY float32 = padding

	if size.Width >= splitThreshold {
		tickerWidth := size.Width * 0.3
		if tickerWidth < minTickerWidth {
			tickerWidth = minTickerWidth
		}

		topBarWidth := size.Width - tickerWidth - 3*padding
		topHeight := fyne.Max(a.topBar.MinSize().Height, a.tickers.MinSize().Height)

		a.tickers.Move(fyne.NewPos(padding, 0))
		a.tickers.Resize(fyne.NewSize(tickerWidth, topHeight))

		topBarHeight := a.topBar.MinSize().Height
		topBarY := (topHeight - topBarHeight) / 2
		a.topBar.Move(fyne.NewPos(tickerWidth+2*padding, topBarY))
		a.topBar.Resize(fyne.NewSize(topBarWidth, topHeight))

		contentY = topHeight
	} else {

		topHeight = a.topBar.MinSize()
		a.topBar.Move(fyne.NewPos(padding, padding))
		a.topBar.Resize(fyne.NewSize(size.Width-2*padding, topHeight.Height))

		contentY += topHeight.Height + padding

		if a.tickers != nil && len(a.tickers.Objects) > 0 {
			tickerHeight = a.tickers.MinSize()

			a.tickers.Move(fyne.NewPos(padding, contentY))
			a.tickers.Resize(fyne.NewSize(size.Width-2*padding, tickerHeight.Height))

			contentY += tickerHeight.Height
		}
	}

	contentHeight := size.Height - contentY - padding

	a.content.Move(fyne.NewPos(padding, contentY))
	a.content.Resize(fyne.NewSize(size.Width-2*padding, contentHeight))

	if a.placeholder != nil {
		a.placeholder.Move(fyne.NewPos(0, -JC.PanelHeight))
	}

	if a.overlay != nil {
		a.overlay.Resize(size)
	}

	JC.MainLayoutContentWidth = size.Width - 2*padding
	JC.MainLayoutContentHeight = contentHeight

	AppLayout.SetMaxOffset(-1)
	AppLayout.SetContentTopY(contentY)
	AppLayout.SetContentBottomY(contentY + contentHeight)
}

func (a *AppMainLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	var top fyne.Size
	if a.topBar != nil {
		top = a.topBar.MinSize()
	}

	var content fyne.Size
	if a.content != nil {
		content = a.content.MinSize()
	}

	width := fyne.Max(top.Width, content.Width) + 2*a.padding
	height := top.Height + content.Height + 3*a.padding

	if a.tickers != nil && len(a.tickers.Objects) > 0 {
		height += a.tickers.MinSize().Height
	}

	return fyne.NewSize(width, height)
}
