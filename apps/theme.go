package apps

import (
	"image/color"
	JC "jxwatcher/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type JXWatcherTheme struct{}

func (t *JXWatcherTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForegroundOnWarning:
		return JC.TextColor
	case theme.ColorNameForegroundOnError:
		return JC.TextColor
	case theme.ColorNameError:
		return JC.ErrorColor // danger background
	}
	return theme.DarkTheme().Color(name, variant)
}

func (t *JXWatcherTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}
func (t *JXWatcherTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}
func (t *JXWatcherTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(name)
}

func NewTheme() fyne.Theme {
	theme := &JXWatcherTheme{}
	return theme
}
