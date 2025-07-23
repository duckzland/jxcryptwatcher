package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	tooltip "github.com/dweymouth/fyne-tooltip/widget"
)

type HoverCursorIconButton struct {
	widget.Button
	tooltip.ToolTipWidgetExtend
}

// NewButtonWithIcon creates a new button widget with the specified label, themed icon, tooltip text and tap handler
// When no icon needed, just pass nil to it
// When no label or tip needed, just pass empty string to it
func NewHoverCursorIconButton(text string, icon fyne.Resource, tip string, onTapped func()) *HoverCursorIconButton {
	b := &HoverCursorIconButton{
		Button: widget.Button{
			Text:     text,
			Icon:     icon,
			OnTapped: onTapped,
		},
	}
	b.ExtendBaseWidget(b)
	if tip != "" {
		b.SetToolTip(tip)
	}
	return b
}

func (b *HoverCursorIconButton) ExtendBaseWidget(wid fyne.Widget) {
	b.ExtendToolTipWidget(wid)
	b.Button.ExtendBaseWidget(wid)
}

func (b *HoverCursorIconButton) MouseIn(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseIn(e)
	b.Button.MouseIn(e)
}

func (b *HoverCursorIconButton) MouseOut() {
	b.ToolTipWidgetExtend.MouseOut()
	b.Button.MouseOut()
}

func (b *HoverCursorIconButton) MouseMoved(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseMoved(e)
	b.Button.MouseMoved(e)
}

// Implement desktop.Cursorable
func (b *HoverCursorIconButton) Cursor() desktop.Cursor {
	if !b.Disabled() {
		return desktop.PointerCursor // Shows hand icon
	}
	return desktop.DefaultCursor
}
