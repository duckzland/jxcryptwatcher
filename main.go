package main

import (
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JC "jxwatcher/core"
)

var signals chan os.Signal

func main() {

	signal.Notify(JC.ShutdownSignal, syscall.SIGINT, syscall.SIGTERM)

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

	sig := <-JC.ShutdownSignal

	JC.Logf("Received signal: %v. Performing cleanup and exiting gracefully.", sig)

	appShutdown()

	signal.Stop(JC.ShutdownSignal)
}
