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
	if t.variant == theme.VariantLight {
		return t.lightColor(name)
	}
	return t.darkColor(name)
}

func (t *appTheme) lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{255, 255, 255, 255}
	case theme.ColorNameButton:
		return color.RGBA{245, 245, 245, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{245, 245, 245, 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{245, 245, 245, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{227, 227, 227, 255}
	case theme.ColorNameInputBorder:
		return color.RGBA{227, 227, 227, 255}
	case theme.ColorNameError:
		return color.RGBA{244, 67, 54, 255}
	case theme.ColorNameFocus:
		return color.RGBA{0, 108, 255, 42}
	case theme.ColorNamePrimary:
		return color.RGBA{0, 108, 255, 42}
	case theme.ColorNameSelection:
		return color.RGBA{0, 108, 255, 64}
	case theme.ColorNameForeground:
		return color.RGBA{13, 13, 13, 255}
	case theme.ColorNameForegroundOnError:
		return color.RGBA{86, 86, 86, 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.RGBA{86, 86, 86, 255}
	case theme.ColorNameForegroundOnSuccess:
		return color.RGBA{86, 86, 86, 255}
	case theme.ColorNameForegroundOnWarning:
		return color.RGBA{86, 86, 86, 255}
	case theme.ColorNameHeaderBackground:
		return color.RGBA{249, 249, 249, 255}
	case theme.ColorNameHover:
		return color.RGBA{0, 0, 0, 15}
	case theme.ColorNameHyperlink:
		return color.RGBA{0, 108, 255, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{243, 243, 243, 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{128, 128, 128, 128}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{136, 136, 136, 255}
	case theme.ColorNamePressed:
		return color.RGBA{0, 0, 0, 25}
	case theme.ColorNameScrollBar:
		return color.RGBA{76, 76, 76, 102}
	case theme.ColorNameScrollBarBackground:
		return color.RGBA{0, 0, 0, 0}
	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 0}
	case theme.ColorNameSeparator:
		return color.RGBA{227, 227, 227, 255}
	case theme.ColorNameSuccess:
		return color.RGBA{67, 244, 54, 255}
	case theme.ColorNameWarning:
		return color.RGBA{255, 152, 0, 255}

	case ColorNamePanelBG:
		return color.RGBA{245, 245, 245, 255}
	case ColorNamePanelPlaceholder:
		return color.RGBA{250, 250, 250, 200}
	case ColorNameTickerBG:
		return color.RGBA{255, 255, 255, 255}
	case ColorNameRed:
		return color.RGBA{255, 180, 180, 255}
	case ColorNameDarkRed:
		return color.RGBA{240, 150, 150, 255}
	case ColorNameGreen:
		return color.RGBA{200, 255, 200, 255}
	case ColorNameDarkGreen:
		return color.RGBA{180, 230, 180, 255}
	case ColorNameBlue:
		return color.RGBA{210, 245, 255, 255}
	case ColorNameLightBlue:
		return color.RGBA{240, 250, 255, 255}
	case ColorNameLightPurple:
		return color.RGBA{250, 240, 255, 255}
	case ColorNameLightOrange:
		return color.RGBA{255, 245, 230, 255}
	case ColorNameOrange:
		return color.RGBA{255, 235, 210, 255}
	case ColorNameYellow:
		return color.RGBA{255, 255, 210, 255}
	case ColorNameTeal:
		return color.RGBA{240, 255, 255, 255}
	case ColorNameDarkGrey:
		return color.RGBA{240, 240, 240, 255}
	case ColorNameTransparent:
		return color.RGBA{0, 0, 0, 0}

	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

func (t *appTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 13, G: 20, B: 33, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 40, G: 41, B: 46, A: 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 40, G: 41, B: 46, A: 255}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 57, G: 57, B: 58, A: 255}
	case theme.ColorNameError:
		return color.RGBA{R: 198, G: 40, B: 40, A: 255}
	case theme.ColorNameFocus:
		return color.RGBA{R: 0, G: 122, B: 204, A: 255}
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
		return color.RGBA{R: 27, G: 27, B: 27, A: 255}
	case theme.ColorNameHover:
		return color.RGBA{R: 15, G: 15, B: 15, A: 15}
	case theme.ColorNameHyperlink:
		return color.RGBA{R: 0, G: 108, B: 255, A: 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 32, G: 32, B: 35, A: 255}
	case theme.ColorNameInputBorder:
		return color.RGBA{R: 57, G: 57, B: 58, A: 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{R: 40, G: 41, B: 46, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 128}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 178, G: 178, B: 178, A: 255}
	case theme.ColorNamePressed:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 122, B: 204, A: 255}
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 50, G: 53, B: 70, A: 255}
	case theme.ColorNameScrollBarBackground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	case theme.ColorNameSelection:
		return color.RGBA{R: 0, G: 122, B: 204, A: 255}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 64, G: 64, B: 64, A: 255}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 67, G: 244, B: 54, A: 255}
	case theme.ColorNameWarning:
		return color.RGBA{R: 255, G: 152, B: 0, A: 255}
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
		return color.RGBA{R: 40, G: 40, B: 40, A: 255}
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
		return 80

	case SizeTickerHeight:
		return 45

	case SizeTickerTitle:
		return 11

	case SizeTickerContent:
		return 18

	case SizeNotificationText:
		return 14

	case SizeCompletionText:
		return 14

	case SizePaddingPanelLeft:
		if IsMobile {
			return 6
		}
		return 8

	case SizePaddingPanelTop:
		if IsMobile {
			return 6
		}
		return 8

	case SizePaddingPanelRight:
		if IsMobile {
			return 6
		}
		return 8

	case SizePaddingPanelBottom:
		if IsMobile {
			return 6
		}
		return 8

	case SizePaddingTickerLeft:
		if IsMobile {
			return 4
		}
		return 8

	case SizePaddingTickerTop:
		if IsMobile {
			return 4
		}
		return 8

	case SizePaddingTickerRight:
		if IsMobile {
			return 4
		}
		return 8

	case SizePaddingTickerBottom:
		if IsMobile {
			return 4
		}
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
