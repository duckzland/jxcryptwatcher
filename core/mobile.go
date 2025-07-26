//go:build mobile
// +build mobile

package core

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const Epsilon = 1e-9
const IsMobile = true
const AppID = "io.github.duckzland.jxcryptwatcher"

var Grid *fyne.Container
var Window fyne.Window
var NotificationBox *widget.Label

var AppBG color.RGBA = color.RGBA{R: 13, G: 20, B: 33, A: 255}
var TextColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var RedColor color.RGBA = color.RGBA{R: 191, G: 8, B: 8, A: 255}
var GreenColor color.RGBA = color.RGBA{R: 2, G: 115, B: 78, A: 255}

var PanelBG color.RGBA = color.RGBA{R: 50, G: 53, B: 70, A: 255}
var PanelBorderRadius float32 = 6
var PanelPadding [4]float32 = [4]float32{0, 5, 10, 5}
var PanelTitleSize float32 = 16
var PanelSubTitleSize float32 = 16
var PanelContentSize float32 = 30
var PanelWidth float32 = 260
var PanelHeight float32 = 180

var UpdateDisplayChan = make(chan struct{})
var UpdateRatesChan = make(chan struct{})
