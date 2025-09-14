package apps

import (
	"sync"

	JW "jxwatcher/widgets"

	"fyne.io/fyne/v2"
)

var AppActionManager *AppActions = &AppActions{}

type AppActions struct {
	mu      sync.RWMutex
	buttons []*JW.HoverCursorIconButton
}

func (a *AppActions) Init() {
	a.mu.Lock()
	a.buttons = []*JW.HoverCursorIconButton{}
	a.mu.Unlock()
}

func (a *AppActions) AddButton(btn *JW.HoverCursorIconButton) {
	a.mu.Lock()
	a.buttons = append(a.buttons, btn)
	a.mu.Unlock()
}

func (a *AppActions) GetButton(tag string) *JW.HoverCursorIconButton {
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
