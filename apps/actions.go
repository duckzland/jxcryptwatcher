package apps

import (
	JW "jxwatcher/widgets"

	"fyne.io/fyne/v2"
)

var AppActionManager *AppMainActions = &AppMainActions{}

type AppMainActions struct {
	Buttons []*JW.HoverCursorIconButton
}

func (a *AppMainActions) Init() {
	a.Buttons = []*JW.HoverCursorIconButton{}
}

func (a *AppMainActions) AddButton(btn *JW.HoverCursorIconButton) {
	a.Buttons = append(a.Buttons, btn)
}

func (a *AppMainActions) GetButton(tag string) *JW.HoverCursorIconButton {
	for _, btn := range a.Buttons {
		if btn.GetTag() == tag {
			return btn
		}
	}

	return nil
}

func (a *AppMainActions) ChangeButtonState(tag string, state string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	fyne.Do(func() {
		btn.ChangeState(state)
	})

	return true
}

func (a *AppMainActions) CallButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	btn.Call()

	return true
}

func (a *AppMainActions) DisableButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	btn.Disable()

	return true
}

func (a *AppMainActions) EnableButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	btn.Enable()

	return true
}

func (a *AppMainActions) DisableAllButton() {
	for _, btn := range a.Buttons {
		btn.Disable()
	}
}

func (a *AppMainActions) EnableAllButton() {
	for _, btn := range a.Buttons {
		btn.Enable()
	}
}
