package apps

import (
	"slices"
	"sync"

	"fyne.io/fyne/v2"

	JW "jxwatcher/widgets"
)

var ActionManager *appActions = &appActions{}

type appActions struct {
	mu      sync.RWMutex
	buttons []*JW.ActionButton
}

func (a *appActions) Init() {
	a.mu.Lock()
	a.buttons = []*JW.ActionButton{}
	a.mu.Unlock()
}

func (a *appActions) Add(btn *JW.ActionButton) {
	a.mu.Lock()
	a.buttons = append(a.buttons, btn)
	a.mu.Unlock()
}

func (a *appActions) Get(tag string) *JW.ActionButton {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, btn := range a.buttons {
		if btn.GetTag() == tag {
			return btn
		}
	}
	return nil
}

func (a *appActions) Call(tag string) bool {
	btn := a.Get(tag)
	if btn == nil {
		return false
	}
	btn.Call()
	return true
}

func (a *appActions) Remove(btn *JW.ActionButton) {
	a.mu.Lock()
	a.buttons = slices.DeleteFunc(a.buttons, func(b *JW.ActionButton) bool {
		return b == btn
	})
	a.mu.Unlock()
}

func (a *appActions) Refresh() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Refresh()
		}
	})
}

func (a *appActions) Disable() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fyne.Do(func() {
		for _, btn := range a.buttons {
			btn.Disable()
		}
	})
}
