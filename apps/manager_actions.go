package apps

import (
	"slices"
	"sync"

	"fyne.io/fyne/v2"

	JW "jxwatcher/widgets"
)

var actionManagerStorage *actionManager = nil

type actionManager struct {
	mu      sync.RWMutex
	buttons []JW.ActionButton
}

func (a *actionManager) Init() {
	a.mu.Lock()
	a.buttons = []JW.ActionButton{}
	a.mu.Unlock()
}

func (a *actionManager) Add(btn JW.ActionButton) {
	a.mu.Lock()
	a.buttons = append(a.buttons, btn)
	a.mu.Unlock()
}

func (a *actionManager) Get(tag string) JW.ActionButton {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, btn := range a.buttons {
		if btn.GetTag() == tag {
			return btn
		}
	}
	return nil
}

func (a *actionManager) Call(tag string) bool {
	btn := a.Get(tag)
	if btn == nil {
		return false
	}
	btn.Call()
	return true
}

func (a *actionManager) Remove(btn JW.ActionButton) {
	a.mu.Lock()
	a.buttons = slices.DeleteFunc(a.buttons, func(b JW.ActionButton) bool {
		return b == btn
	})
	a.mu.Unlock()
}

func (a *actionManager) Refresh() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Refresh()
		}
	})
}

func (a *actionManager) Disable() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Disable()
		}
	})
}

func RegisterActionManager() *actionManager {
	if actionManagerStorage == nil {
		actionManagerStorage = &actionManager{}
	}
	return actionManagerStorage
}

func UseAction() *actionManager {
	return actionManagerStorage
}
