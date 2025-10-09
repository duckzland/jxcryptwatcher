package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type dialogFormTheme struct {
	base fyne.Theme
}

func (t *dialogFormTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameScrollBar:
		return 1
	case theme.SizeNameScrollBarSmall:
		return 1
	default:
		return t.base.Size(name)
	}
}

func (t *dialogFormTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameScrollBar || name == theme.ColorNameScrollBarBackground {
		return color.RGBA{0, 0, 0, 0}
	}
	return t.base.Color(name, variant)
}
func (t *dialogFormTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *dialogFormTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}
