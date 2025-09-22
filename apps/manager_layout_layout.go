package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
)

type mainLayout struct {
	padding     float32
	parent      *layoutManager
	topBar      *fyne.Container
	tickers     *fyne.Container
	overlay     *fyne.Container
	content     *container.Scroll
	placeholder *canvas.Rectangle
}

func (a *mainLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
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

		ntPos := fyne.NewPos(padding, 0)
		if a.tickers.Position() != ntPos {
			a.tickers.Move(ntPos)
		}

		ntSize := fyne.NewSize(tickerWidth, topHeight)
		if a.tickers.Size() != ntSize {
			a.tickers.Resize(ntSize)
		}

		topBarHeight := a.topBar.MinSize().Height
		topBarY := (topHeight - topBarHeight) / 2

		tpPos := fyne.NewPos(tickerWidth+2*padding, topBarY)
		if a.topBar.Position() != tpPos {
			a.topBar.Move(tpPos)
		}

		tpSize := fyne.NewSize(topBarWidth, topHeight)
		if a.topBar.Size() != tpSize {
			a.topBar.Resize(tpSize)
		}

		contentY = topHeight
	} else {

		topHeight = a.topBar.MinSize()
		topBarSize := fyne.NewSize(size.Width-2*padding, topHeight.Height)
		topBarPos := fyne.NewPos(padding, padding)

		if a.topBar.Size() != topBarSize {
			a.topBar.Resize(topBarSize)
		}

		if a.topBar.Position() != topBarPos {
			a.topBar.Move(topBarPos)
		}

		contentY += topHeight.Height + padding

		if a.tickers != nil && len(a.tickers.Objects) > 0 {
			tickerHeight = a.tickers.MinSize()

			ntPos := fyne.NewPos(padding, contentY)
			if a.tickers.Position() != ntPos {
				a.tickers.Move(ntPos)
			}

			ntSize := fyne.NewSize(size.Width-2*padding, tickerHeight.Height)
			if a.tickers.Size() != ntSize {
				a.tickers.Resize(ntSize)
			}

			contentY += tickerHeight.Height
		}
	}

	contentHeight := size.Height - contentY - padding

	ctPos := fyne.NewPos(padding, contentY)
	if a.content.Position() != ctPos {
		a.content.Move(ctPos)
	}

	ctSize := fyne.NewSize(size.Width-2*padding, contentHeight)
	if a.content.Size() != ctSize {
		a.content.Resize(ctSize)
	}

	if a.placeholder != nil {
		pPos := fyne.NewPos(0, -JC.PanelHeight)

		if a.placeholder.Position() != pPos {
			a.placeholder.Move(pPos)
		}
	}

	if a.overlay != nil {
		if a.overlay.Size() != size {
			a.overlay.Resize(size)
		}
	}

	JC.MainLayoutContentWidth = size.Width - 2*padding
	JC.MainLayoutContentHeight = contentHeight

	if a.parent != nil {
		a.parent.SetMaxOffset(-1)
		a.parent.SetContentTopY(contentY)
		a.parent.SetContentBottomY(contentY + contentHeight)
	}
}

func (a *mainLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
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
