package widgets

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"

	JC "jxwatcher/core"
)

var ActiveEntry *CompletionEntry = nil

type CompletionEntry struct {
	widget.Entry
	popup            *fyne.Container
	container        *fyne.Container
	navigableList    *navigableList
	parent           *ExtendedFormDialog
	pause            bool
	itemHeight       float32
	options          []string
	suggestions      []string
	lowerSuggestions []string
	popupPosition    fyne.Position
	canvas           fyne.Canvas
	lastInput        string
	uuid             string
}

func NewCompletionEntry(
	options []string,
	popup *fyne.Container,
) *CompletionEntry {

	c := &CompletionEntry{suggestions: options, popup: popup}
	c.ExtendBaseWidget(c)

	id := uuid.New()
	c.uuid = id.String()

	c.OnChanged = c.SearchSuggestions

	c.popupPosition = fyne.NewPos(-1, -1)

	c.itemHeight = 36

	c.options = options

	c.suggestions = options

	c.navigableList = NewNavigableList(
		c.setTextFromMenu,
		c.HideCompletion,
	)

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		c.HideCompletion()
	})

	closeBtn.Resize(fyne.NewSize(32, 32))
	closeBtn.Move(fyne.NewPos(0, 0))

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	bg.CornerRadius = JC.PanelBorderRadius

	c.container = container.NewStack(
		bg,
		container.New(
			&CompletionListEntryLayout{},
			c.navigableList,
			closeBtn,
		),
	)

	c.popup.Add(c.container)

	c.popup.Hide()

	// Caching the suggestions, this is needed for faster search!
	go func() {
		c.lowerSuggestions = make([]string, len(c.suggestions))
		for i, s := range c.suggestions {
			c.lowerSuggestions[i] = strings.ToLower(s)
		}
	}()

	return c
}

func (c *CompletionEntry) GetCurrentInput() string {
	return c.lastInput
}

func (c *CompletionEntry) SearchSuggestions(s string) {

	if s == c.lastInput {
		return
	}

	c.lastInput = s

	if c.pause {
		JC.MainDebouncer.Cancel("show_suggestion_" + c.uuid)
		return
	}

	delay := 10 * time.Millisecond
	if c.popup.Visible() {
		delay = 50 * time.Millisecond
	}

	minText := 1

	// Bail out early
	if len(s) < minText || s == "" {
		JC.MainDebouncer.Cancel("show_suggestion_" + c.uuid)
		fyne.Do(func() {
			c.HideCompletion()
		})
		return
	}

	JC.MainDebouncer.Call("show_suggestion_"+c.uuid, delay, func() {

		input := c.GetCurrentInput()

		if len(input) < minText || input == "" {
			fyne.Do(func() {
				c.HideCompletion()
			})
			return
		}

		lowerS := strings.ToLower(input)
		results := []string{}

		for i, part := range c.lowerSuggestions {
			if strings.Contains(part, lowerS) {
				display := c.suggestions[i]
				results = append(results, display)
			}
		}

		if len(results) == 0 {
			fyne.Do(func() {
				c.HideCompletion()
			})
			return
		}

		results = JC.ReorderByMatch(results, input)

		fyne.Do(func() {
			c.SetOptions(results)
			c.ShowCompletion()
			c.DynamicResize()
		})
	})
}

func (c *CompletionEntry) TypedKey(event *fyne.KeyEvent) {
	c.Entry.TypedKey(event)
}

func (c *CompletionEntry) FocusLost() {

	c.Entry.FocusLost()

	if JC.IsMobile {

		// Fix for when android keyboard hiding, position got bad
		JC.MainDebouncer.Call("completion_entry_positioning_"+c.uuid, 100*time.Millisecond, func() {
			fyne.Do(func() {
				if c.popup.Visible() {
					c.popup.Move(c.popUpPos())
					canvas.Refresh(c.popup)
				}
			})
		})
	}
}

func (c *CompletionEntry) FocusGained() {
	c.Entry.FocusGained()

	if ActiveEntry != nil && ActiveEntry != c {
		ActiveEntry.HideCompletion()
	}

	if len(c.Text) > 0 {
		c.navigableList.SetFilteredData(c.options)
		c.ShowCompletion()
	}
}

func (c *CompletionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *CompletionEntry) HideCompletion() {
	if c.popup != nil {
		c.popup.Hide()
	}

	if c.navigableList != nil {
		c.navigableList.SetFilteredData([]string{})
	}

	c.popupPosition = fyne.NewPos(-1, -1)
}

func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
}

func (c *CompletionEntry) Resize(size fyne.Size) {
	c.Entry.Resize(size)

	JC.MainDebouncer.Call("completion_resizing_"+c.uuid, 50*time.Millisecond, func() {
		fyne.Do(func() {
			if c.popup != nil && c.popup.Visible() {
				c.popupPosition = fyne.NewPos(-1, -1)
				c.popup.Resize(c.maxSize())
				c.popup.Move(c.popUpPos())
			}
		})
	})

}

func (c *CompletionEntry) DynamicResize() {
	mx := c.maxSize()
	ox := c.popup.Size()

	if mx.Width != ox.Width || mx.Height != ox.Height {
		c.popup.Resize(mx)
		canvas.Refresh(c.popup)
	}
}

func (c *CompletionEntry) SetOptions(itemList []string) {

	c.options = itemList

	if c.navigableList != nil {
		c.navigableList.SetFilteredData(c.options)
	}
}

func (c *CompletionEntry) SetParent(parent *ExtendedFormDialog) {
	c.parent = parent
}

func (c *CompletionEntry) ShowCompletion() {
	if c.pause {
		return
	}

	if len(c.options) == 0 || len(c.Text) == 0 {
		c.HideCompletion()
		return
	}

	if c.popup.Visible() && c.popupPosition.Y != -1 {
		return
	}

	// Always reset position cache!
	c.calculatePosition(true)

	c.navigableList.UnselectAll()
	c.navigableList.selected = -1

	mx := c.maxSize()
	ox := c.popup.Size()

	mp := c.popUpPos()
	op := c.popup.Position()

	refresh := false

	if mx.Width != ox.Width {
		c.popup.Resize(mx)
		refresh = true
	}

	if mp.X != op.X || mp.Y != op.Y {
		c.popup.Move(c.popUpPos())
		refresh = true
	}

	if len(c.popup.Objects) == 0 {
		c.popup.Add(c.container)
	}

	if refresh {
		canvas.Refresh(c.popup)
	}

	c.popup.Show()

	ActiveEntry = c
}

func (c *CompletionEntry) calculatePosition(force bool) bool {

	if c.canvas == nil {
		c.canvas = fyne.CurrentApp().Driver().CanvasForObject(c)
	}

	if c.canvas == nil {
		return false
	}

	if c.parent == nil || c.parent.overlayContent == nil {
		return false
	}

	if c.popupPosition.Y != -1 && !force {
		return true
	}

	p := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	x := fyne.CurrentApp().Driver().AbsolutePositionForObject(c.parent.overlayContent)
	px := p.Subtract(x)

	c.popupPosition = px

	return true
}

func (c *CompletionEntry) maxSize() fyne.Size {

	if !c.calculatePosition(false) {
		return fyne.NewSize(0, 0)
	}

	if c.canvas == nil {
		return fyne.NewSize(0, 0)
	}

	padding := (theme.Padding() * 2) * c.canvas.Scale()
	listHeight := float32(len(c.options)) * (c.itemHeight)
	maxHeight := c.canvas.Size().Height - c.popupPosition.Y - c.Size().Height - padding

	if maxHeight > 300 {
		maxHeight = 300
	}

	if JC.IsMobile {
		maxHeight = 200
	}

	if listHeight > maxHeight {
		listHeight = maxHeight
	}

	return fyne.NewSize(c.Size().Width, listHeight)
}

func (c *CompletionEntry) popUpPos() fyne.Position {
	if !c.calculatePosition(false) {
		return fyne.NewPos(0, 0)
	}

	entryPos := c.popupPosition
	entryPos.Y += c.Size().Height
	entryPos.Y += 2

	return entryPos

}

func (c *CompletionEntry) setTextFromMenu(s string) {
	JC.MainDebouncer.Cancel("show_suggestion_" + c.uuid)

	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.popup.Hide()
	c.pause = false
}

type CompletionListEntryLayout struct {
	cSize fyne.Size
}

func (l *CompletionListEntryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	if size == l.cSize {
		return
	}

	l.cSize = size

	listEntry := objects[0]
	closeBtn := objects[1]

	height := size.Height
	closeWidth := closeBtn.Size().Width

	closeBtn.Resize(fyne.NewSize(closeWidth, closeWidth))
	closeBtn.Move(fyne.NewPos(-closeWidth-2, 2))

	listEntry.Resize(fyne.NewSize(size.Width, height))
	listEntry.Move(fyne.NewPos(0, 0))
}

func (l *CompletionListEntryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}

	listEntry := objects[0]
	closeBtn := objects[1]

	listMin := listEntry.MinSize()
	closeMin := closeBtn.MinSize()

	width := listMin.Width + closeMin.Width
	height := fyne.Max(listMin.Height, closeMin.Height)

	l.cSize = fyne.NewSize(width, height)

	return l.cSize
}
