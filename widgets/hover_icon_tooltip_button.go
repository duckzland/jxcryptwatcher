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
	tag           string
	state         string
	disabled      bool
	allow_actions bool
	validate      func(*HoverCursorIconButton)
}

func NewHoverCursorIconButton(
	tag string,
	text string,
	icon fyne.Resource,
	tip string,
	state string,
	onTapped func(*HoverCursorIconButton),
	validate func(*HoverCursorIconButton),
) *HoverCursorIconButton {

	b := &HoverCursorIconButton{
		Button: widget.Button{
			Text:       text,
			Icon:       icon,
			Importance: widget.MediumImportance,
		},
		tag:      tag,
		disabled: false,
		validate: validate,
	}

	b.setState(state)

	b.Button.OnTapped = func() {
		if !b.disabled && b.allow_actions {
			onTapped(b)
		}
	}

	// Only extend tooltip if tip is non-empty
	if tip != "" {
		b.ExtendToolTipWidget(b)
		b.SetToolTip(tip)
	}

	b.Button.ExtendBaseWidget(b)

	b.allow_actions = true

	return b
}

func (b *HoverCursorIconButton) ExtendBaseWidget(wid fyne.Widget) {
	b.Button.ExtendBaseWidget(wid)
}

func (b *HoverCursorIconButton) MouseIn(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseIn(e)

	if !b.disabled {
		b.Button.MouseIn(e)
	}
}

func (b *HoverCursorIconButton) MouseOut() {
	if !b.allow_actions {
		b.ToolTipWidgetExtend.MouseOut()
		return
	}

	b.ToolTipWidgetExtend.MouseOut()

	if !b.disabled {
		b.Button.MouseOut()
	}
}

func (b *HoverCursorIconButton) MouseMoved(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseMoved(e)

	if !b.disabled {
		b.Button.MouseMoved(e)
	}
}

func (b *HoverCursorIconButton) Tapped(_ *fyne.PointEvent) {
	if !b.allow_actions {
		return
	}

	if b.disabled {
		return
	}
	b.Button.Tapped(nil)
}

func (b *HoverCursorIconButton) Cursor() desktop.Cursor {
	if !b.disabled && b.allow_actions {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (b *HoverCursorIconButton) DisallowActions() {
	b.changeState("disallow_actions")
}

func (b *HoverCursorIconButton) AllowActions() {
	b.changeState("allow_actions")
}

func (b *HoverCursorIconButton) Disable() {
	b.changeState("disabled")
}

func (b *HoverCursorIconButton) Enable() {
	b.changeState("reset")
}

func (b *HoverCursorIconButton) Error() {
	b.changeState("error")
}

func (b *HoverCursorIconButton) Progress() {
	b.changeState("in_progress")
}

func (b *HoverCursorIconButton) Active() {
	b.changeState("active")
}

func (b *HoverCursorIconButton) GetTag() string {
	return b.tag
}

func (b *HoverCursorIconButton) Refresh() {
	if b.validate != nil {
		b.validate(b)
	}
	fyne.Do(b.Button.Refresh)
}

func (b *HoverCursorIconButton) Call() {
	if !b.allow_actions {
		return
	}

	b.Button.OnTapped()
}

func (b *HoverCursorIconButton) setState(state string) {

	if b.state == state {
		return
	}

	switch state {
	case "allow_actions":
		b.allow_actions = true
		b.Button.Importance = widget.MediumImportance
	case "disallow_actions":
		b.ToolTipWidgetExtend.MouseOut()
		b.Button.MouseOut()
		b.Button.Importance = widget.MediumImportance
		b.allow_actions = false
	case "disabled":
		b.allow_actions = true
		b.disabled = true
		b.Button.Importance = widget.LowImportance
	case "in_progress":
		b.allow_actions = true
		b.disabled = true
		b.Button.Importance = widget.HighImportance
	case "active":
		b.allow_actions = true
		b.Button.Importance = widget.HighImportance
	case "error":
		b.allow_actions = true
		b.disabled = false
		b.Button.Importance = widget.DangerImportance
	case "reset", "normal":
		b.allow_actions = true
		b.disabled = false
		b.Button.Importance = widget.MediumImportance
	}

	b.state = state
}

func (b *HoverCursorIconButton) changeState(state string) {

	if b.state == state {
		return
	}

	b.setState(state)
	fyne.Do(b.Button.Refresh)
}
