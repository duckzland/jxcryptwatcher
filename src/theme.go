package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomDarkTheme provides a fresh dark theme, with no fallback to Fyne defaults
type CustomDarkTheme struct{}

func (CustomDarkTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 38, G: 38, B: 38, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 130, G: 130, B: 130, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 28, G: 28, B: 28, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 90, G: 60, B: 255, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	default:
		return color.Black
	}
}

func (CustomDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Use built-in fonts, or add your custom font here
	if style.Bold {
		return theme.DefaultTextBoldFont()
	} else if style.Italic {
		return theme.DefaultTextItalicFont()
	}
	return theme.DefaultTextFont()
}

func (CustomDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// Use default icons, or replace with your own
	return theme.DefaultTheme().Icon(name)
}

func (CustomDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 6
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	default:
		return 14
	}
}
