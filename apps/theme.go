package apps

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type JXWatcherTheme struct{}

func (t *JXWatcherTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForegroundOnWarning:
		return color.White
	case theme.ColorNameForegroundOnError:
		return color.White
	case theme.ColorNameError:
		return color.RGBA{R: 220, G: 20, B: 60, A: 255} // danger background
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
