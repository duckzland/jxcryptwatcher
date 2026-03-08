package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BannerVariant int

const (
	BannerDanger BannerVariant = iota
	BannerWarning
	BannerAction
)

type Banner struct {
	widget.BaseWidget
	text    string
	variant BannerVariant
}

func (b *Banner) CreateRenderer() fyne.WidgetRenderer {
	var bgColor, fgColor fyne.ThemeColorName
	switch b.variant {
	case BannerDanger:
		bgColor = theme.ColorNameError
		fgColor = theme.ColorNameForegroundOnError
	case BannerWarning:
		bgColor = theme.ColorNameWarning
		fgColor = theme.ColorNameForegroundOnWarning
	case BannerAction:
		bgColor = theme.ColorNamePrimary
		fgColor = theme.ColorNameForegroundOnPrimary
	}

	bg := canvas.NewRectangle(theme.Color(bgColor))
	bg.CornerRadius = theme.Size(theme.SizeNameInputRadius)

	lbl := widget.NewRichText(
		&widget.TextSegment{
			Text: b.text,
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
				ColorName: fgColor,
			},
		},
	)

	return widget.NewSimpleRenderer(container.NewStack(bg, lbl))
}

func NewBanner(msg string, variant BannerVariant) *Banner {
	b := &Banner{text: msg, variant: variant}
	b.ExtendBaseWidget(b)
	return b
}
