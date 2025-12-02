package animations

import "image/color"

type AnimatableText interface {
	Visible() bool
	Hide()
	Show()
	Refresh()
	SetText(string)
	SetAlpha(uint8)
	SetColor(color.Color)
}
