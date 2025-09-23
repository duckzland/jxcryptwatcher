package core

import (
	"image/color"

	"fyne.io/fyne/v2"
)

var App fyne.App
var Window fyne.Window

var Tickers *fyne.Container

var ThemeColor func(name fyne.ThemeColorName) color.Color
var ThemeSize func(name fyne.ThemeSizeName) float32
