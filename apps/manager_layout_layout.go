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
		topHeight := fyne.NewSize(topBarWidth, fyne.Max(a.topBar.MinSize().Height, a.tickers.MinSize().Height))

		tickerSize := fyne.NewSize(tickerWidth, topHeight.Height)
		tickerPos := fyne.NewPos(padding, 0)

		if a.tickers.Size() != tickerSize {
			a.tickers.Resize(tickerSize)
		}

		if a.tickers.Position() != tickerPos {
			a.tickers.Move(tickerPos)
		}

		topBarHeight := a.topBar.MinSize().Height
		topBarY := (topHeight.Height - topBarHeight) / 2
		topBarSize := fyne.NewSize(topBarWidth, topHeight.Height)
		topBarPos := fyne.NewPos(tickerWidth+2*padding, topBarY)

		if a.topBar.Size() != topBarSize {
			a.topBar.Resize(topBarSize)
		}

		if a.topBar.Position() != topBarPos {
			a.topBar.Move(topBarPos)
		}

		contentY = topHeight.Height
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
			tickerSize := fyne.NewSize(size.Width-2*padding, tickerHeight.Height)
			tickerPos := fyne.NewPos(padding, contentY)

			if a.tickers.Size() != tickerSize {
				a.tickers.Resize(tickerSize)
			}

			if a.tickers.Position() != tickerPos {
				a.tickers.Move(tickerPos)
			}
			contentY += tickerHeight.Height
		}
	}

	contentSize := fyne.NewSize(size.Width-2*padding, size.Height-contentY-padding)
	contentPos := fyne.NewPos(padding, contentY)
	if a.content.Size() != contentSize {
		a.content.Resize(contentSize)
	}

	if a.content.Position() != contentPos {
		a.content.Move(contentPos)
	}

	if a.placeholder != nil {
		placeholderPos := fyne.NewPos(0, -JC.PanelHeight)
		if a.placeholder.Position() != placeholderPos {
			a.placeholder.Move(placeholderPos)
		}
	}

	if a.overlay != nil {
		if a.overlay.Size() != size {
			a.overlay.Resize(size)
		}
	}

	JC.MainLayoutContentWidth = contentSize.Width
	JC.MainLayoutContentHeight = contentSize.Height

	if a.parent != nil {
		a.parent.SetMaxOffset(-1)
		a.parent.SetContentTopY(contentY)
		a.parent.SetContentBottomY(contentY + contentSize.Height)
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
