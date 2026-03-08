package widgets

import (
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type selectEntry struct {
	widget.Select
	action    func(active bool)
	options   map[int]string
	Validator func(int, string) error
}

func (e *selectEntry) SetAction(fn func(active bool)) {
	e.action = fn
}

func (e *selectEntry) FocusGained() {
	e.Select.FocusGained()
	if e.action != nil {
		e.action(true)
	}
}

func (e *selectEntry) FocusLost() {
	e.Select.FocusLost()
	if e.action != nil {
		e.action(false)
	}
}

func (e *selectEntry) Keyboard() mobile.KeyboardType {
	return mobile.DefaultKeyboard
}

func (e *selectEntry) SetDefaultValue(val int) {
	if s, ok := e.options[val]; ok {
		e.Selected = s
	}
}

func (e *selectEntry) SetValidator(fn func(int, string) error) {
	e.Validator = fn
}

func (e *selectEntry) Validate() error {
	if e.Validator != nil {
		return e.Validator(e.GetInt(), e.Selected)
	}
	return nil
}

func (e *selectEntry) GetInt() int {
	for k, v := range e.options {
		if v == e.Selected {
			return k
		}
	}
	return -1
}

func NewSelectEntry(options map[int]string) *selectEntry {
	entry := &selectEntry{options: options}
	entry.Select.Options = make([]string, 0, len(options))
	for _, v := range options {
		entry.Select.Options = append(entry.Select.Options, v)
	}
	entry.Select.OnChanged = func(s string) {
		if entry.Validator != nil {
			_ = entry.Validator(entry.GetInt(), s)
		}
	}
	entry.ExtendBaseWidget(entry)
	return entry
}
