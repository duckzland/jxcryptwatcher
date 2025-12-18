package core

import (
	"context"

	"fyne.io/fyne/v2"
)

var App fyne.App
var Window fyne.Window
var ShutdownCtx, ShutdownCancel = context.WithCancel(context.Background())
var AppInFocus bool
