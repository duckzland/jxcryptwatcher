package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	tooltip "github.com/dweymouth/fyne-tooltip/widget"

	JC "jxwatcher/core"
)

const ActionStateAllowActions = "allow_actions"
const ActionStateDisallowActions = "disallow_actions"
const ActionStateDisabled = "disabled"
const ActionStateInProgress = "in_progress"
const ActionStateActive = "active"
const ActionStateError = "error"
const ActionStateReset = "reset"
const ActionStateNormal = "normal"

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
	Destroy()
	Refresh()
	GetTag() string
}

type actionButton struct {
	widget.BaseWidget
	tooltip.ToolTipWidgetExtend
	tag           string
	state         string
	disabled      bool
	allow_actions bool
	hastip        bool
	validate      func(ActionButton)
	buttonWidget  fyne.Widget
}

func (b *actionButton) ExtendBaseWidget(wid fyne.Widget) {
	switch btn := b.buttonWidget.(type) {
	case *widget.Button:
		btn.ExtendBaseWidget(wid)
	case *actionButtonIcon:
		btn.ExtendBaseWidget(wid)
	default:
	}
}

func (b *actionButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.buttonWidget)
}

func (b *actionButton) MinSize() fyne.Size {
	return b.buttonWidget.MinSize()
}

func (b *actionButton) Resize(size fyne.Size) {
	b.BaseWidget.Resize(size)
	b.buttonWidget.Resize(size)
	b.buttonWidget.Move(fyne.NewPos(0, 0))

}

func (b *actionButton) MouseIn(e *desktop.MouseEvent) {
	if !b.allow_actions || b.disabled {
		if b.hastip {
			b.ToolTipWidgetExtend.MouseOut()
		}
		return
	}

	if b.hastip {
		b.ToolTipWidgetExtend.MouseIn(e)
	}

	if !b.disabled {
		if hoverable, ok := b.buttonWidget.(desktop.Hoverable); ok {
			hoverable.MouseIn(e)
		}
	}
}

func (b *actionButton) MouseOut() {
	if b.hastip {
		b.ToolTipWidgetExtend.MouseOut()
	}

	if !b.disabled {
		if hoverable, ok := b.buttonWidget.(desktop.Hoverable); ok {
			hoverable.MouseOut()
		}
	}
}

func (b *actionButton) MouseMoved(e *desktop.MouseEvent) {
	if !b.allow_actions {
		if b.hastip {
			b.ToolTipWidgetExtend.MouseOut()
		}
		return
	}

	if b.hastip {
		b.ToolTipWidgetExtend.MouseMoved(e)
	}

	if !b.disabled {
		if hoverable, ok := b.buttonWidget.(desktop.Hoverable); ok {
			hoverable.MouseMoved(e)
		}
	}
}

func (b *actionButton) Tapped(e *fyne.PointEvent) {
	if !b.allow_actions || b.disabled {
		if b.hastip {
			b.ToolTipWidgetExtend.MouseOut()
		}
		return
	}

	if b.hastip {
		b.ToolTipWidgetExtend.MouseOut()
	}

	if tappable, ok := b.buttonWidget.(fyne.Tappable); ok {
		tappable.Tapped(e)
	}
}

func (b *actionButton) Cursor() desktop.Cursor {
	if !b.disabled && b.allow_actions {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (b *actionButton) IsDisabled() bool {
	return b.disabled
}

func (b *actionButton) DisallowActions() {
	b.changeState(ActionStateDisallowActions)
}

func (b *actionButton) AllowActions() {
	b.changeState(ActionStateAllowActions)
}

func (b *actionButton) Disable() {
	b.changeState(ActionStateDisabled)
}

func (b *actionButton) Enable() {
	b.changeState(ActionStateReset)
}

func (b *actionButton) Error() {
	b.changeState(ActionStateError)
}

func (b *actionButton) Progress() {
	b.changeState(ActionStateInProgress)
}

func (b *actionButton) Active() {
	b.changeState(ActionStateActive)
}

func (b *actionButton) GetTag() string {
	return b.tag
}

func (b *actionButton) Refresh() {
	if b.validate != nil {
		b.validate(b)
	}
	fyne.Do(b.buttonWidget.Refresh)
}

func (b *actionButton) Destroy() {
	if b.hastip {
		b.ToolTipWidgetExtend.MouseOut()
	}
}

func (b *actionButton) Call() {
	if !b.allow_actions {
		return
	}

	switch btn := b.buttonWidget.(type) {
	case *widget.Button:
		if btn.OnTapped != nil {
			btn.OnTapped()
		}
	case *actionButtonIcon:
		if btn.onTapped != nil {
			btn.onTapped()
		}
	}
}

func (b *actionButton) setImportance(i widget.Importance) {
	switch btn := b.buttonWidget.(type) {
	case *widget.Button:
		btn.Importance = i
	case *actionButtonIcon:
		btn.Importance = i
	}
}

func (b *actionButton) triggerMouseOut() {
	if b.hastip {
		b.ToolTipWidgetExtend.MouseOut()
	}
	if hoverable, ok := b.buttonWidget.(desktop.Hoverable); ok {
		hoverable.MouseOut()
	}
}

func (b *actionButton) setState(state string) {
	if b.state == state {
		return
	}

	switch state {
	case ActionStateAllowActions:
		b.allow_actions = true
		b.setImportance(widget.MediumImportance)

	case ActionStateDisallowActions:
		b.allow_actions = false
		b.triggerMouseOut()
		b.setImportance(widget.MediumImportance)

	case ActionStateDisabled:
		b.allow_actions = true
		b.disabled = true
		b.setImportance(widget.LowImportance)

	case ActionStateInProgress:
		b.allow_actions = true
		b.disabled = true
		b.setImportance(widget.HighImportance)

	case ActionStateActive:
		b.allow_actions = true
		b.setImportance(widget.HighImportance)

	case ActionStateError:
		b.allow_actions = true
		b.disabled = false
		b.setImportance(widget.DangerImportance)

	case ActionStateReset, ActionStateNormal:
		b.allow_actions = true
		b.disabled = false
		b.setImportance(widget.MediumImportance)
	}

	b.state = state
}

func (b *actionButton) changeState(state string) {

	if b.state == state {
		return
	}

	b.setState(state)
	fyne.Do(b.buttonWidget.Refresh)
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
		tag:           tag,
		disabled:      false,
		hastip:        false,
		allow_actions: true,
		validate:      validate,
	}

	cb := func() {
		if b.disabled {
			return
		}
		onTapped(b)
	}

	if text == JC.STRING_EMPTY && icon != nil {
		b.buttonWidget = NewActionButtonIcon(icon, widget.MediumImportance, cb)
	} else {
		b.buttonWidget = widget.NewButtonWithIcon(text, icon, cb)
	}

	if tip != JC.STRING_EMPTY {
		b.ExtendToolTipWidget(b)
		b.SetToolTip(tip)
		b.hastip = true
	}

	b.setState(state)

	b.ExtendBaseWidget(b.buttonWidget)

	return b
}
