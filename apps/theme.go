package apps

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JC "jxwatcher/core"
)

var AppTheme *appTheme

type appTheme struct{}

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x39, G: 0x39, B: 0x3a, A: 0xff}
	case theme.ColorNameError:
		return JC.ErrorColor
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x00, G: 0x7a, B: 0xcc, A: 0xff}
	case theme.ColorNameForeground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameForegroundOnError:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameForegroundOnSuccess:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameForegroundOnWarning:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x1b, G: 0x1b, B: 0x1b, A: 0xff}
	case theme.ColorNameHover:
		return color.RGBA{R: 30, G: 30, B: 30, A: 255}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x00, G: 0x6c, B: 0xff, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x20, G: 0x20, B: 0x23, A: 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x39, G: 0x39, B: 0x3a, A: 0xff}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 128}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x00, G: 0x7a, B: 0xcc, A: 0xff}
	case theme.ColorNameScrollBar:
		return JC.PanelBG
	case theme.ColorNameScrollBarBackground:
		return JC.Transparent
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x00, G: 0x7a, B: 0xcc, A: 0xff}
	case theme.ColorNameSeparator:
		return color.Gray{Y: 64}
	case theme.ColorNameShadow:
		return JC.Transparent
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff}
	default:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}
	}
}

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTextFont()
}

func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameLineSpacing:
		return 4
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 12
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 5
	case theme.SizeNameSelectionRadius:
		return 3
	case theme.SizeNameScrollBarRadius:
		return 3
	case theme.SizeNameWindowButtonHeight:
		return 16
	case theme.SizeNameWindowButtonRadius:
		return 8
	case theme.SizeNameWindowButtonIcon:
		return 14
	case theme.SizeNameWindowTitleBarHeight:
		return 26
	default:
		return 0
	}
}

func NewTheme() fyne.Theme {
	return &appTheme{}
}
