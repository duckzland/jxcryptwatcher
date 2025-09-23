package core

import (
	"image/color"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
)

var App fyne.App
var Window fyne.Window
var HWTotalCPU = runtime.NumCPU()

var PanelBorderRadius float32 = 6
var PanelPadding [4]float32 = [4]float32{8, 8, 8, 8}

var PanelTitleSize float32 = 16
var PanelSubTitleSize float32 = 16
var PanelBottomTextSize float32 = 12
var PanelContentSize float32 = 28
var PanelTitleSizeSmall float32 = 13
var PanelSubTitleSizeSmall float32 = 13
var PanelBottomTextSizeSmall float32 = 10
var PanelContentSizeSmall float32 = 22
var PanelWidth float32 = 320
var PanelHeight float32 = 110
var ActionBtnWidth float32 = 40
var ActionBtnGap float32 = 6

var TickerBorderRadius float32 = 6
var TickerPadding [4]float32 = [4]float32{8, 8, 8, 8}

var TickerWidth float32 = 120
var TickerHeight float32 = 50
var TickerTitleSize float32 = 11
var TickerContentSize float32 = 20

var NotificationTextSize float32 = 14

var CompletionTextSize float32 = 14

var UpdateStatusChan = make(chan string, 1000)
var UpdateDisplayTimestamp = time.Now()

var MainDebouncer = NewDebouncer()

var MainLayoutContentWidth float32
var MainLayoutContentHeight float32

var Tickers *fyne.Container

var AnimDispatcher = NewDispatcher(100, 4, 16*time.Millisecond)

var CharWidthCache = make(map[int]float32)

var ThemeColor func(name fyne.ThemeColorName) color.Color
