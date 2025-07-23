package core

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var Grid *fyne.Container
var Window fyne.Window
var NotificationBox *widget.Label

// @todo Move these to theme
var AppBG color.RGBA = color.RGBA{R: 13, G: 20, B: 33, A: 255}
var PanelBG color.RGBA = color.RGBA{R: 50, G: 53, B: 70, A: 255}
var TextColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var RedColor color.RGBA = color.RGBA{R: 191, G: 8, B: 8, A: 255}
var GreenColor color.RGBA = color.RGBA{R: 2, G: 115, B: 78, A: 255}

const Epsilon = 1e-9
