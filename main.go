package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JC "jxwatcher/core"
)

func main() {

	JC.App = app.NewWithID(JC.AppID)

	JC.Window = JC.App.NewWindow("JXCrypto Watcher")

	registerLogger()

	registerTheme()

	registerAppIcon()

	registerFonts()

	registerUtility()

	registerCache()

	registerActions()

	registerFetchers()

	registerWorkers()

	registerDispatcher()

	registerShutdown()

	registerLifecycle()

	JC.Window.Resize(fyne.NewSize(920, 600))

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Window.ShowAndRun()
}
