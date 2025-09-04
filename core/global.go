package core

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const Epsilon = 1e-9

var Window fyne.Window

var AppBG color.RGBA = color.RGBA{R: 13, G: 20, B: 33, A: 255}
var TextColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var Transparent color.RGBA = color.RGBA{R: 0, G: 0, B: 0, A: 0}
var RedColor = color.RGBA{R: 220, G: 60, B: 60, A: 255}
var GreenColor = color.RGBA{R: 28, G: 132, B: 87, A: 255}
var BlueColor = color.RGBA{R: 60, G: 120, B: 220, A: 255}
var LightBlueColor = color.RGBA{R: 100, G: 160, B: 230, A: 255}
var LightPurpleColor = color.RGBA{R: 160, G: 140, B: 200, A: 255}
var LightOrangeColor = color.RGBA{R: 240, G: 160, B: 100, A: 255}
var OrangeColor = color.RGBA{R: 230, G: 120, B: 60, A: 255}
var YellowColor = color.RGBA{R: 240, G: 210, B: 80, A: 255}
var TealGreenColor = color.RGBA{R: 40, G: 170, B: 140, A: 255}

var PanelBG color.RGBA = color.RGBA{R: 50, G: 53, B: 70, A: 255}
var PanelPlaceholderBG color.RGBA = color.RGBA{R: 20, G: 22, B: 30, A: 200}
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

var TickerBG color.RGBA = color.RGBA{R: 17, G: 119, B: 170, A: 255}
var TickerBorderRadius float32 = 6
var TickerPadding [4]float32 = [4]float32{8, 8, 8, 8}

var TickerWidth float32 = 120
var TickerHeight float32 = 50
var TickerTitleSize float32 = 11
var TickerContentSize float32 = 20

var UpdateStatusChan = make(chan string, 1000)
var UpdateDisplayChan = make(chan struct{})
var UpdateRatesChan = make(chan struct{})
var UpdateTickersChan = make(chan struct{})
var UpdateDisplayTimestamp = time.Now()

var MainDebouncer = NewDebouncer()

var MainLayoutContentWidth float32
var MainLayoutContentHeight float32

var NotificationContainer *canvas.Rectangle = nil

var Tickers *fyne.Container
