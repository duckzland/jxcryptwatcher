package main

import (
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JC "jxwatcher/core"
)

func main() {

	JC.App = app.NewWithID(JC.AppID)

	JC.Window = JC.App.NewWindow("JXCrypto Watcher")

	registerBoot()

	registerTheme()

	registerAppIcon()

	registerFonts()

	registerUtility()

	registerCache()

	registerActions()

	registerFetchers()

	registerWorkers()

	registerDispatcher()

	registerLifecycle()

	JC.Window.Resize(fyne.NewSize(920, 600))

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Window.ShowAndRun()

	JC.Logln("Received exit signal, Performing cleanup and exiting gracefully.")

	timer := time.NewTimer(2 * time.Second)
	go func() {
		<-timer.C
		JC.Logln("Force exiting after timeout...")
		os.Exit(1)
	}()

	appShutdown()

	<-JC.ShutdownCtx.Done()

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}

	}
}
