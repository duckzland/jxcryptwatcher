package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type textEntry struct {
	widget.Entry
	action func(active bool)
}

func (e *textEntry) SetAction(fn func(active bool)) {
	e.action = fn
}

func (e *textEntry) FocusGained() {
	e.Entry.FocusGained()
	if e.action != nil {
		e.action(true)
	}
}

func (e *textEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.action != nil {
		e.action(false)
	}
}

func (e *textEntry) TypedShortcut(shortcut fyne.Shortcut) {
	e.Entry.TypedShortcut(shortcut)
}

func (e *textEntry) Keyboard() mobile.KeyboardType {
	return mobile.DefaultKeyboard
}

func (e *textEntry) SetDefaultValue(s string) {
	e.Text = s
}

func (e *textEntry) SetValidator(fn func(string) error) {
	e.Validator = fn
}

func NewTextEntry() *textEntry {
	entry := &textEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}
