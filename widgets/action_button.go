package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	tooltip "github.com/dweymouth/fyne-tooltip/widget"
)

type ActionButton interface {
	fyne.Widget
	desktop.Hoverable
	Show()
	Hide()
	IsDisabled() bool
	Disable()
	Enable()
	DisallowActions()
	AllowActions()
	Error()
	Progress()
	Active()
	Call()
	Refresh()
	GetTag() string
}

type actionButton struct {
	widget.Button
	tooltip.ToolTipWidgetExtend
	tag           string
	state         string
	disabled      bool
	allow_actions bool
	validate      func(ActionButton)
}

func NewActionButton(
	tag string,
	text string,
	icon fyne.Resource,
	tip string,
	state string,
	onTapped func(ActionButton),
	validate func(ActionButton),
) ActionButton {

	b := &actionButton{
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

func (b *actionButton) ExtendBaseWidget(wid fyne.Widget) {
	b.Button.ExtendBaseWidget(wid)
}

func (b *actionButton) MouseIn(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseIn(e)

	if !b.disabled {
		b.Button.MouseIn(e)
	}
}

func (b *actionButton) MouseOut() {
	if !b.allow_actions {
		b.ToolTipWidgetExtend.MouseOut()
		return
	}

	b.ToolTipWidgetExtend.MouseOut()

	if !b.disabled {
		b.Button.MouseOut()
	}
}

func (b *actionButton) MouseMoved(e *desktop.MouseEvent) {
	if !b.allow_actions {
		return
	}

	b.ToolTipWidgetExtend.MouseMoved(e)

	if !b.disabled {
		b.Button.MouseMoved(e)
	}
}

func (b *actionButton) Tapped(_ *fyne.PointEvent) {
	if !b.allow_actions {
		return
	}

	if b.disabled {
		return
	}
	b.Button.Tapped(nil)
}

func (b *actionButton) Cursor() desktop.Cursor {
	if !b.disabled && b.allow_actions {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (b *actionButton) IsDisabled() bool {
	return b.Disabled()
}

func (b *actionButton) DisallowActions() {
	b.changeState("disallow_actions")
}

func (b *actionButton) AllowActions() {
	b.changeState("allow_actions")
}

func (b *actionButton) Disable() {
	b.changeState("disabled")
}

func (b *actionButton) Enable() {
	b.changeState("reset")
}

func (b *actionButton) Error() {
	b.changeState("error")
}

func (b *actionButton) Progress() {
	b.changeState("in_progress")
}

func (b *actionButton) Active() {
	b.changeState("active")
}

func (b *actionButton) GetTag() string {
	return b.tag
}

func (b *actionButton) Refresh() {
	if b.validate != nil {
		b.validate(b)
	}
	fyne.Do(b.Button.Refresh)
}

func (b *actionButton) Call() {
	if !b.allow_actions {
		return
	}

	b.Button.OnTapped()
}

func (b *actionButton) setState(state string) {

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

func (b *actionButton) changeState(state string) {

	if b.state == state {
		return
	}

	b.setState(state)
	fyne.Do(b.Button.Refresh)
}
