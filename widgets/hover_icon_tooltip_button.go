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
	tag      string
	state    string
	disabled bool
	validate func(*HoverCursorIconButton)
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
	}

	b.tag = tag
	b.disabled = false
	b.validate = validate

	b.setState(state)

	b.Button.OnTapped = func() {
		if b.disabled == false {
			onTapped(b)
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

	if !b.disabled {
		b.Button.MouseIn(e)
	}
}

func (b *HoverCursorIconButton) MouseOut() {
	b.ToolTipWidgetExtend.MouseOut()

	if !b.disabled {
		b.Button.MouseOut()
	}
}

func (b *HoverCursorIconButton) MouseMoved(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseMoved(e)
	if !b.disabled {
		b.Button.MouseMoved(e)
	}
}

func (b *HoverCursorIconButton) Tapped(_ *fyne.PointEvent) {
	if b.disabled {
		return
	}
	b.Button.Tapped(nil)
}

func (b *HoverCursorIconButton) Cursor() desktop.Cursor {
	if !b.disabled {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
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
	b.Button.OnTapped()
}

func (b *HoverCursorIconButton) setState(state string) {

	if b.state == state {
		return
	}

	switch state {
	case "disabled":
		b.disabled = true
		b.Button.Importance = widget.LowImportance
	case "in_progress":
		b.disabled = true
		b.Button.Importance = widget.HighImportance
	case "active":
		b.Button.Importance = widget.HighImportance
	case "error":
		b.disabled = false
		b.Button.Importance = widget.DangerImportance
	case "reset", "normal":
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
