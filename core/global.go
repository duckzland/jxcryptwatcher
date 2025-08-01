package core

import (
	"image/color"

	"fyne.io/fyne/v2"
)

const Epsilon = 1e-9

var Grid *fyne.Container
var Window fyne.Window

var AppBG color.RGBA = color.RGBA{R: 13, G: 20, B: 33, A: 255}
var TextColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var RedColor color.RGBA = color.RGBA{R: 191, G: 8, B: 8, A: 255}
var GreenColor color.RGBA = color.RGBA{R: 2, G: 115, B: 78, A: 255}

var PanelBG color.RGBA = color.RGBA{R: 50, G: 53, B: 70, A: 255}
var PanelBorderRadius float32 = 6
var PanelPadding [4]float32 = [4]float32{20, 8, 0, 8}
var PanelTitleSize float32 = 16
var PanelSubTitleSize float32 = 16
var PanelContentSize float32 = 30
var PanelWidth float32 = 320
var PanelHeight float32 = 130
var ActionBtnWidth float32 = 40
var ActionBtnGap float32 = 6

var UpdateStatusChan = make(chan string, 1000)
var UpdateDisplayChan = make(chan struct{})
var UpdateRatesChan = make(chan struct{})

var MainLayoutContentWidth float32
var MainLayoutContentHeight float32
