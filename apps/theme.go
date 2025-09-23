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
	case theme.ColorNameBackground:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x39, G: 0x39, B: 0x3a, A: 0xff}
	case theme.ColorNameError:
		return color.RGBA{R: 198, G: 40, B: 40, A: 255}
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
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f}
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
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}
	case theme.ColorNameScrollBarBackground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x00, G: 0x7a, B: 0xcc, A: 0xff}
	case theme.ColorNameSeparator:
		return color.Gray{Y: 64}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff}

	case JC.ColorNamePanelBG:
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}
	case JC.ColorNamePanelPlaceholder:
		return color.RGBA{R: 20, G: 22, B: 30, A: 200}
	case JC.ColorNameTickerBG:
		return color.RGBA{R: 17, G: 119, B: 170, A: 255}
	case JC.ColorNameRed:
		return color.RGBA{R: 133, G: 36, B: 36, A: 255}
	case JC.ColorNameGreen:
		return color.RGBA{R: 22, G: 106, B: 69, A: 255}
	case JC.ColorNameBlue:
		return color.RGBA{R: 60, G: 120, B: 220, A: 255}
	case JC.ColorNameLightBlue:
		return color.RGBA{R: 100, G: 160, B: 230, A: 255}
	case JC.ColorNameLightPurple:
		return color.RGBA{R: 160, G: 140, B: 200, A: 255}
	case JC.ColorNameLightOrange:
		return color.RGBA{R: 240, G: 160, B: 100, A: 255}
	case JC.ColorNameOrange:
		return color.RGBA{R: 195, G: 102, B: 51, A: 255}
	case JC.ColorNameYellow:
		return color.RGBA{R: 192, G: 168, B: 64, A: 255}
	case JC.ColorNameTeal:
		return color.RGBA{R: 40, G: 170, B: 140, A: 255}
	case JC.ColorNameTransparent:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}

	default:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}
	}
}

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(fyne.TextStyle{Monospace: true})
	}
	if style.Bold && style.Italic {
		return theme.DefaultTheme().Font(fyne.TextStyle{Bold: true, Italic: true})
	}
	if style.Bold {
		return theme.DefaultTheme().Font(fyne.TextStyle{Bold: true})
	}
	if style.Italic {
		return theme.DefaultTheme().Font(fyne.TextStyle{Italic: true})
	}
	return theme.DefaultTheme().Font(fyne.TextStyle{})
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
		if JC.IsMobile {
			return 4
		}
		return 12
	case theme.SizeNameScrollBarSmall:
		if JC.IsMobile {
			return 4
		}
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

	// Injected custom sizes with JC. prefix
	case JC.SizePanelBorderRadius:
		return 6
	case JC.SizePanelTitle:
		return 16
	case JC.SizePanelSubTitle:
		return 16
	case JC.SizePanelBottomText:
		return 12
	case JC.SizePanelContent:
		return 28
	case JC.SizePanelTitleSmall:
		return 13
	case JC.SizePanelSubTitleSmall:
		return 13
	case JC.SizePanelBottomTextSmall:
		return 10
	case JC.SizePanelContentSmall:
		return 22
	case JC.SizePanelWidth:
		return 320
	case JC.SizePanelHeight:
		return 110
	case JC.SizeActionBtnWidth:
		return 40
	case JC.SizeActionBtnGap:
		return 6
	case JC.SizeTickerBorderRadius:
		return 6
	case JC.SizeTickerWidth:
		return 120
	case JC.SizeTickerHeight:
		return 50
	case JC.SizeTickerTitle:
		return 11
	case JC.SizeTickerContent:
		return 20
	case JC.SizeNotificationText:
		return 14
	case JC.SizeCompletionText:
		return 14

	case JC.SizePaddingPanelLeft:
		return 8
	case JC.SizePaddingPanelTop:
		return 8
	case JC.SizePaddingPanelRight:
		return 8
	case JC.SizePaddingPanelBottom:
		return 8

	case JC.SizePaddingTickerLeft:
		return 8
	case JC.SizePaddingTickerTop:
		return 8
	case JC.SizePaddingTickerRight:
		return 8
	case JC.SizePaddingTickerBottom:
		return 8

	default:
		return 0
	}
}

var activeTheme = &appTheme{}

func NewTheme() fyne.Theme {
	if JC.ThemeColor == nil {
		JC.ThemeColor = func(name fyne.ThemeColorName) color.Color {
			return activeTheme.Color(name, theme.VariantDark)
		}
	}

	if JC.ThemeSize == nil {
		JC.ThemeSize = func(name fyne.ThemeSizeName) float32 {
			return activeTheme.Size(name)
		}
	}

	return activeTheme
}
