package apps

import (
	"fyne.io/fyne/v2"

	JW "jxwatcher/widgets"
)

var AppActionManager *AppActions = &AppActions{}

type AppActions struct {
	Buttons []*JW.HoverCursorIconButton
}

func (a *AppActions) Init() {
	a.Buttons = []*JW.HoverCursorIconButton{}
}

func (a *AppActions) AddButton(btn *JW.HoverCursorIconButton) {
	a.Buttons = append(a.Buttons, btn)
}

func (a *AppActions) GetButton(tag string) *JW.HoverCursorIconButton {
	for _, btn := range a.Buttons {
		if btn.GetTag() == tag {
			return btn
		}
	}

	return nil
}

func (a *AppActions) CallButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	btn.Call()

	return true
}

func (a *AppActions) Refresh() {
	fyne.Do(func() {
		for _, btn := range a.Buttons {
			btn.Refresh()
		}
	})
}

func (a *AppActions) Disable() {
	fyne.Do(func() {
		for _, btn := range a.Buttons {
			btn.Disable()
		}
	})
}
