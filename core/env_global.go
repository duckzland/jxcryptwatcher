package core

import (
	"context"
	"os"

	"fyne.io/fyne/v2"
)

var App fyne.App
var Window fyne.Window
var ShutdownSignal chan os.Signal = make(chan os.Signal, 1)
var ShutdownCtx, ShutdownCancel = context.WithCancel(context.Background())
