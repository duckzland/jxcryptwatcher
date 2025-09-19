package apps

import (
	"slices"
	"sync"

	"fyne.io/fyne/v2"

	JW "jxwatcher/widgets"
)

var AppActionManager *AppActions = &AppActions{}

type AppActions struct {
	mu      sync.RWMutex
	buttons []*JW.ActionButton
}

func (a *AppActions) Init() {
	a.mu.Lock()
	a.buttons = []*JW.ActionButton{}
	a.mu.Unlock()
}

func (a *AppActions) AddButton(btn *JW.ActionButton) {
	a.mu.Lock()
	a.buttons = append(a.buttons, btn)
	a.mu.Unlock()
}

func (a *AppActions) GetButton(tag string) *JW.ActionButton {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, btn := range a.buttons {
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

func (a *AppActions) RemoveButton(btn *JW.ActionButton) {
	a.mu.Lock()
	a.buttons = slices.DeleteFunc(a.buttons, func(b *JW.ActionButton) bool {
		return b == btn
	})
	a.mu.Unlock()
}

func (a *AppActions) Refresh() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Refresh()
		}
	})
}

func (a *AppActions) Disable() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Disable()
		}
	})
}
