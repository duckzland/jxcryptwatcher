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

var UpdateStatusChan = make(chan string, 1000)
var UpdateDisplayTimestamp = time.Now()

var MainDebouncer = NewDebouncer()

var MainLayoutContentWidth float32
var MainLayoutContentHeight float32

var Tickers *fyne.Container

var AnimDispatcher = NewDispatcher(100, 4, 16*time.Millisecond)

var CharWidthCache = make(map[int]float32)

var ThemeColor func(name fyne.ThemeColorName) color.Color
var ThemeSize func(name fyne.ThemeSizeName) float32
