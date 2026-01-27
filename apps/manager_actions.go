package apps

import (
	"sync"

	JW "jxwatcher/widgets"
)

var actionManagerStorage *actionManager = nil

type actionManager struct {
	buttons sync.Map // key: string (tag), value: JW.ActionButton
}

func (a *actionManager) Init() {
	a.buttons = sync.Map{}
}

func (a *actionManager) Add(btn JW.ActionButton) {
	if btn == nil {
		return
	}
	a.buttons.Store(btn.GetTag(), btn)
}

func (a *actionManager) Get(tag string) JW.ActionButton {
	v, ok := a.buttons.Load(tag)
	if !ok {
		return nil
	}
	return v.(JW.ActionButton)
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
	if btn == nil {
		return
	}
	a.buttons.Delete(btn.GetTag())
}

func (a *actionManager) Refresh() {
	a.buttons.Range(func(_, v any) bool {
		if v != nil {
			v.(JW.ActionButton).Refresh()
		}
		return true
	})
}

func (a *actionManager) Disable() {
	a.buttons.Range(func(_, v any) bool {
		if v != nil {
			v.(JW.ActionButton).Disable()
		}
		return true
	})
}

func (a *actionManager) HideTooltip() {
	a.buttons.Range(func(_, v any) bool {
		if v != nil {
			v.(JW.ActionButton).HideTooltip()
		}
		return true
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
