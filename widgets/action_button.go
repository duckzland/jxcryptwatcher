package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	tooltip "github.com/dweymouth/fyne-tooltip/widget"
)

type ActionButton struct {
	widget.Button
	tooltip.ToolTipWidgetExtend
	tag           string
	state         string
	disabled      bool
	allow_actions bool
	validate      func(*ActionButton)
}

func NewActionButton(
	tag string,
	text string,
	icon fyne.Resource,
	tip string,
	state string,
	onTapped func(*ActionButton),
	validate func(*ActionButton),
) *ActionButton {

	b := &ActionButton{
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

func (b *ActionButton) ExtendBaseWidget(wid fyne.Widget) {
	b.Button.ExtendBaseWidget(wid)
}

func (b *ActionButton) MouseIn(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseIn(e)

	if !b.disabled {
		b.Button.MouseIn(e)
	}
}

func (b *ActionButton) MouseOut() {
	if !b.allow_actions {
		b.ToolTipWidgetExtend.MouseOut()
		return
	}

	b.ToolTipWidgetExtend.MouseOut()

	if !b.disabled {
		b.Button.MouseOut()
	}
}

func (b *ActionButton) MouseMoved(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseMoved(e)

	if !b.disabled {
		b.Button.MouseMoved(e)
	}
}

func (b *ActionButton) Tapped(_ *fyne.PointEvent) {
	if !b.allow_actions {
		return
	}

	if b.disabled {
		return
	}
	b.Button.Tapped(nil)
}

func (b *ActionButton) Cursor() desktop.Cursor {
	if !b.disabled && b.allow_actions {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (b *ActionButton) DisallowActions() {
	b.changeState("disallow_actions")
}

func (b *ActionButton) AllowActions() {
	b.changeState("allow_actions")
}

func (b *ActionButton) Disable() {
	b.changeState("disabled")
}

func (b *ActionButton) Enable() {
	b.changeState("reset")
}

func (b *ActionButton) Error() {
	b.changeState("error")
}

func (b *ActionButton) Progress() {
	b.changeState("in_progress")
}

func (b *ActionButton) Active() {
	b.changeState("active")
}

func (b *ActionButton) GetTag() string {
	return b.tag
}

func (b *ActionButton) Refresh() {
	if b.validate != nil {
		b.validate(b)
	}
	fyne.Do(b.Button.Refresh)
}

func (b *ActionButton) Call() {
	if !b.allow_actions {
		return
	}

	b.Button.OnTapped()
}

func (b *ActionButton) setState(state string) {

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

func (b *ActionButton) changeState(state string) {

	if b.state == state {
		return
	}

	b.setState(state)
	fyne.Do(b.Button.Refresh)
}
