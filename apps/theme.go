package apps

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JC "jxwatcher/core"
)

type appTheme struct{}

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
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

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}
func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}
func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(name)
}

func NewTheme() fyne.Theme {
	theme := &appTheme{}
	return theme
}
