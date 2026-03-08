package widgets

import (
	"fyne.io/fyne/v2/widget"
)

type radioEntry struct {
	widget.RadioGroup
	options map[int]string
}

func (e *radioEntry) SetDefaultValue(val int) {
	if s, ok := e.options[val]; ok {
		e.Selected = s
		e.Refresh()
	}
}

func (e *radioEntry) GetInt() int {
	for k, v := range e.options {
		if v == e.Selected {
			return k
		}
	}
	return -1
}

func NewRadioEntry(options map[int]string, onChanged func(string)) *radioEntry {
	entry := &radioEntry{options: options}
	entry.RadioGroup.Options = make([]string, 0, len(options))

	for _, v := range options {
		entry.RadioGroup.Options = append(entry.RadioGroup.Options, v)
	}

	//entry.RadioGroup.Horizontal = true
	entry.RadioGroup.OnChanged = onChanged

	entry.ExtendBaseWidget(entry)

	return entry
}
