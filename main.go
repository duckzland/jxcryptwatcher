package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JC "jxwatcher/core"
)

func main() {
	JC.InitLogger()

	JC.Logln("App is booting...")

	JC.RegisterThemeManager().Init()

	JC.App = app.NewWithID(JC.AppID)
	// Comment this out for now, as we dont have real settings to force DarkTheme
	// JC.UseTheme().SetVariant(JC.App.Settings().ThemeVariant())

	JC.App.Settings().SetTheme(JC.UseTheme())

	JC.Window = JC.App.NewWindow("JXCrypto Watcher")

	setAppIcon()

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
