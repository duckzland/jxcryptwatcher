package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type actionButtonIcon struct {
	widget.BaseWidget
	widget.DisableableWidget
	Icon       fyne.Resource
	Importance widget.Importance
	Disabled   bool
	onTapped   func()
	focused    bool
}

func (b *actionButtonIcon) Tapped(_ *fyne.PointEvent) {
	if b.Disabled {
		return
	}
	if b.onTapped != nil {
		b.onTapped()
	}
}

func (b *actionButtonIcon) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	icon := widget.NewIcon(b.Icon)

	r := &actionButtonIconRenderer{
		button:     b,
		background: bg,
		icon:       icon,
		objects:    []fyne.CanvasObject{bg, icon},
	}
	r.applyTheme()
	return r
}

type actionButtonIconRenderer struct {
	button     *actionButtonIcon
	background *canvas.Rectangle
	icon       *widget.Icon
	objects    []fyne.CanvasObject
}

func (r *actionButtonIconRenderer) Layout(size fyne.Size) {
	th := r.button.Theme()

	iconSize := fyne.NewSquareSize(th.Size(theme.SizeNameInlineIcon))
	iconPos := fyne.NewPos(
		(size.Width-iconSize.Width)/2,
		(size.Height-iconSize.Height)/2,
	)

	if iconSize != r.icon.Size() {
		r.icon.Resize(iconSize)
	}

	if iconPos != r.icon.Position() {
		r.icon.Move(iconPos)
	}

	if size != r.background.Size() {
		r.background.Resize(size)
	}
}

func (r *actionButtonIconRenderer) MinSize() fyne.Size {
	pad := r.padding()
	iconSize := theme.IconInlineSize()
	return fyne.NewSize(iconSize+pad.Width, iconSize+pad.Height)
}

func (r *actionButtonIconRenderer) Refresh() {
	r.applyTheme()
	canvas.Refresh(r.button)
}

func (r *actionButtonIconRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *actionButtonIconRenderer) Destroy() {}

func (r *actionButtonIconRenderer) applyTheme() {
	th := r.button.Theme()

	fg, bg, blend := r.buttonColorNames()

	bgColor := theme.Color(bg)

	if blend != "" {
		bgColor = r.blendColor(bgColor, theme.Color(blend))
	}

	if r.background.FillColor != bgColor {
		r.background.FillColor = bgColor
		r.background.CornerRadius = th.Size(theme.SizeNameInputRadius)
		r.background.Refresh()
	}

	if r.icon.Resource != nil {
		icon := r.icon.Resource
		if thRes, ok := icon.(fyne.ThemedResource); ok {
			if thRes.ThemeColorName() != fg {
				icon = theme.NewColoredResource(icon, fg)
			}
		}

		if r.icon.Resource != icon {
			r.icon.Resource = icon
			r.icon.Refresh()
		}
	}
}

func (r *actionButtonIconRenderer) buttonColorNames() (fg, bg, blend fyne.ThemeColorName) {
	fg = theme.ColorNameForeground
	b := r.button

	if b.Disabled {
		fg = theme.ColorNameDisabled
		if b.Importance != widget.LowImportance {
			bg = theme.ColorNameDisabledButton
		}
	} else if b.focused {
		blend = theme.ColorNameFocus
	}

	if bg == "" {
		switch b.Importance {
		case widget.DangerImportance:
			fg = theme.ColorNameForegroundOnError
			bg = theme.ColorNameError
		case widget.HighImportance:
			fg = theme.ColorNameForegroundOnPrimary
			bg = theme.ColorNamePrimary
		case widget.LowImportance:
			if blend != "" {
				bg = theme.ColorNameButton
			}
		case widget.SuccessImportance:
			fg = theme.ColorNameForegroundOnSuccess
			bg = theme.ColorNameSuccess
		case widget.WarningImportance:
			fg = theme.ColorNameForegroundOnWarning
			bg = theme.ColorNameWarning
		default:
			bg = theme.ColorNameButton
		}
	}
	return fg, bg, blend
}

func (r *actionButtonIconRenderer) padding() fyne.Size {
	return fyne.NewSquareSize(r.button.Theme().Size(theme.SizeNameInnerPadding) * 2)
}

func (r *actionButtonIconRenderer) blendColor(base, overlay color.Color) color.Color {
	br, bg, bb, ba := base.RGBA()
	or, og, ob, oa := overlay.RGBA()

	return &color.NRGBA{
		R: uint8((br*(65535-oa) + or*oa) / 65535 >> 8),
		G: uint8((bg*(65535-oa) + og*oa) / 65535 >> 8),
		B: uint8((bb*(65535-oa) + ob*oa) / 65535 >> 8),
		A: uint8(ba >> 8),
	}
}

func NewActionButtonIcon(icon fyne.Resource, importance widget.Importance, onTapped func()) *actionButtonIcon {
	b := &actionButtonIcon{
		Icon:       icon,
		Importance: importance,
		onTapped:   onTapped,
	}
	b.ExtendBaseWidget(b)
	return b
}
