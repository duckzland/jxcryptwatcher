package main

import (
	"syscall"

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

	signals <- syscall.SIGTERM
	<-finalShutdown
}
