package apps

import (
	JC "jxwatcher/core"
	JW "jxwatcher/widgets"

	"fyne.io/fyne/v2"
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

func (a *AppActions) ChangeButtonState(tag string, state string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	fyne.Do(func() {
		btn.ChangeState(state)
	})

	return true
}

func (a *AppActions) CallButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	btn.Call()

	return true
}

func (a *AppActions) DisableButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	fyne.Do(btn.Disable)

	return true
}

func (a *AppActions) EnableButton(tag string) bool {
	btn := a.GetButton(tag)
	if btn == nil {
		return false
	}

	fyne.Do(btn.Enable)

	return true
}

func (a *AppActions) DisableAllButton(exclude string) {
	for _, btn := range a.Buttons {
		if btn.GetTag() == exclude {
			continue
		}
		JC.Logln("Disabling button", btn.GetTag(), exclude)
		fyne.Do(btn.Disable)
	}
}

func (a *AppActions) EnableAllButton(exclude string) {
	for _, btn := range a.Buttons {
		if btn.GetTag() == exclude {
			continue
		}
		fyne.Do(btn.Enable)
	}
}

func (a *AppActions) Refresh() {
	for _, btn := range a.Buttons {
		// if btn.GetTag() == exclude {
		// 	continue
		// }
		fyne.Do(btn.Refresh)
	}
}
