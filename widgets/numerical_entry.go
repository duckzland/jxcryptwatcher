package widgets

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type numericalEntry struct {
	widget.Entry
	allow_decimals bool
}

func NewNumericalEntry(allow_decimals bool) *numericalEntry {
	entry := &numericalEntry{}
	entry.ExtendBaseWidget(entry)
	entry.allow_decimals = allow_decimals
	return entry
}

func (e *numericalEntry) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
	} else if e.allow_decimals && r == '.' && !strings.Contains(e.Text, ".") {
		e.Entry.TypedRune(r)
	}
}

func (e *numericalEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}

	content := paste.Clipboard.Content()
	if _, err := strconv.ParseFloat(content, 64); err == nil {
		e.Entry.TypedShortcut(shortcut)
	}
}

func (e *numericalEntry) Keyboard() mobile.KeyboardType {
	return mobile.DefaultKeyboard

	// Fyne has bug, on android the dot key cant be pressed!
	// return mobile.NumberKeyboard
}

func (e *numericalEntry) SetDefaultValue(s string) {
	e.Text = s
}
