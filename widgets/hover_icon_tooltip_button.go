package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	tooltip "github.com/dweymouth/fyne-tooltip/widget"
)

type HoverCursorIconButton struct {
	tag string
	widget.Button
	tooltip.ToolTipWidgetExtend
	disabled bool
}

func NewHoverCursorIconButton(
	tag string,
	text string,
	icon fyne.Resource,
	tip string,
	onTapped func(btn *widget.Button),
) *HoverCursorIconButton {

	b := &HoverCursorIconButton{
		Button: widget.Button{
			Text:       text,
			Icon:       icon,
			Importance: widget.MediumImportance,
		},
	}

	b.tag = tag
	b.disabled = false

	b.Button.OnTapped = func() {
		if b.disabled == false {
			onTapped(&b.Button)
		}
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

func (b *HoverCursorIconButton) Cursor() desktop.Cursor {
	if !b.Disabled() {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (b *HoverCursorIconButton) Disable() {
	b.disabled = true
}

func (b *HoverCursorIconButton) Enable() {
	b.disabled = false
}

func (b *HoverCursorIconButton) GetTag() string {
	return b.tag
}

func (b *HoverCursorIconButton) ChangeState(state string) {
	switch state {
	case "disabled":
		b.disabled = true
		b.Button.Importance = widget.LowImportance
	case "in_progress":
		b.disabled = true
		b.Button.Importance = widget.HighImportance
	case "error":
		b.disabled = false
		b.Button.Importance = widget.DangerImportance
	case "reset":
		b.disabled = false
		b.Button.Importance = widget.MediumImportance
	}

	b.Button.Refresh()
}

func (b *HoverCursorIconButton) Call() {
	b.Button.OnTapped()
}
