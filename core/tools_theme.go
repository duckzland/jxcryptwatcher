package core

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var activeTheme *appTheme = nil

type appTheme struct {
	mu      sync.Mutex
	variant fyne.ThemeVariant
}

func (t *appTheme) Init() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.variant = theme.VariantDark
}

func (t *appTheme) SetVariant(variant fyne.ThemeVariant) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.variant = variant
}

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {

	case theme.ColorNameBackground:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}

	case theme.ColorNameButton:
		return color.NRGBA{R: 40, G: 41, B: 46, A: 255}

	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 40, G: 41, B: 46, A: 255}

	case theme.ColorNameDisabled:
		return color.NRGBA{R: 57, G: 57, B: 58, A: 255}

	case theme.ColorNameError:
		return color.RGBA{R: 198, G: 40, B: 40, A: 255}

	case theme.ColorNameFocus:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 255}

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
		return color.NRGBA{R: 27, G: 27, B: 27, A: 255}

	case theme.ColorNameHover:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 15}

	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0, G: 108, B: 255, A: 255}

	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 32, G: 32, B: 35, A: 255}

	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 57, G: 57, B: 58, A: 255}

	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 40, G: 41, B: 46, A: 255}

	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 128}

	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 178, G: 178, B: 178, A: 255}

	case theme.ColorNamePressed:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 0}

	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 255}

	case theme.ColorNameScrollBar:
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}

	case theme.ColorNameScrollBarBackground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}

	case theme.ColorNameSelection:
		return color.NRGBA{R: 0, G: 122, B: 204, A: 255}

	case theme.ColorNameSeparator:
		return color.Gray{Y: 64}

	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}

	case theme.ColorNameSuccess:
		return color.NRGBA{R: 67, G: 244, B: 54, A: 255}

	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 152, B: 0, A: 255}

	case ColorNamePanelBG:
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}

	case ColorNamePanelPlaceholder:
		return color.RGBA{R: 20, G: 22, B: 30, A: 200}

	case ColorNameTickerBG:
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}

	case ColorNameRed:
		return color.RGBA{R: 133, G: 36, B: 36, A: 255}

	case ColorNameDarkRed:
		return color.RGBA{R: 100, G: 25, B: 25, A: 255}

	case ColorNameGreen:
		return color.RGBA{R: 22, G: 106, B: 69, A: 255}

	case ColorNameDarkGreen:
		return color.RGBA{R: 15, G: 70, B: 45, A: 255}

	case ColorNameBlue:
		return color.RGBA{R: 60, G: 120, B: 220, A: 255}

	case ColorNameLightBlue:
		return color.RGBA{R: 100, G: 160, B: 230, A: 255}

	case ColorNameLightPurple:
		return color.RGBA{R: 160, G: 140, B: 200, A: 255}

	case ColorNameLightOrange:
		return color.RGBA{R: 240, G: 160, B: 100, A: 255}

	case ColorNameOrange:
		return color.RGBA{R: 195, G: 102, B: 51, A: 255}

	case ColorNameYellow:
		return color.RGBA{R: 192, G: 168, B: 64, A: 255}

	case ColorNameTeal:
		return color.RGBA{R: 40, G: 170, B: 140, A: 255}

	case ColorNameDarkGrey:
		return color.RGBA{R: 40, G: 40, B: 40, A: 255} // Dark Grey

	case ColorNameTransparent:
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
		if IsMobile {
			return 4
		}
		return 12

	case theme.SizeNameScrollBarSmall:
		if IsMobile {
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

	case SizePanelBorderRadius:
		return 6

	case SizePanelTitle:
		return 16

	case SizePanelSubTitle:
		return 16

	case SizePanelBottomText:
		return 12

	case SizePanelContent:
		return 26

	case SizePanelTitleSmall:
		return 13

	case SizePanelSubTitleSmall:
		return 13

	case SizePanelBottomTextSmall:
		return 10

	case SizePanelContentSmall:
		return 22

	case SizePanelWidth:
		return 320

	case SizePanelHeight:
		return 110

	case SizeActionBtnWidth:
		return 40

	case SizeActionBtnGap:
		return 6

	case SizeTickerBorderRadius:
		return 6

	case SizeTickerWidth:
		return 120

	case SizeTickerHeight:
		return 50

	case SizeTickerTitle:
		return 11

	case SizeTickerContent:
		return 18

	case SizeNotificationText:
		return 14

	case SizeCompletionText:
		return 14

	case SizePaddingPanelLeft:
		return 8

	case SizePaddingPanelTop:
		return 8

	case SizePaddingPanelRight:
		return 8

	case SizePaddingPanelBottom:
		return 8

	case SizePaddingTickerLeft:
		return 8

	case SizePaddingTickerTop:
		return 8

	case SizePaddingTickerRight:
		return 8

	case SizePaddingTickerBottom:
		return 8

	default:
		return 0
	}
}

func (t *appTheme) GetColor(name fyne.ThemeColorName) color.Color {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Color(name, t.variant)
}

func RegisterThemeManager() *appTheme {
	if activeTheme == nil {
		activeTheme = &appTheme{}
	}
	return activeTheme
}

func UseTheme() *appTheme {
	return activeTheme
}
