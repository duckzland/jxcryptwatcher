package widgets

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

var ActiveEntry *CompletionEntry = nil

type CompletionEntry struct {
	widget.Entry
	popup            *fyne.Container
	container        *fyne.Container
	navigableList    *navigableList
	pause            bool
	itemHeight       float32
	Parent           *ExtendedFormDialog
	Options          []string
	Suggestions      []string
	CustomCreate     func() fyne.CanvasObject
	CustomUpdate     func(id widget.ListItemID, object fyne.CanvasObject)
	PopupPosition    fyne.Position
	EntryHeight      float32
	EntryWidth       float32
	Canvas           fyne.Canvas
	Scale            float32
	optionsHash      string
	newHash          string
	lowerSuggestions []string
	lastInput        string
}

func NewCompletionEntry(
	options []string,
	popup *fyne.Container,
) *CompletionEntry {

	c := &CompletionEntry{Suggestions: options, popup: popup}
	c.ExtendBaseWidget(c)

	c.OnChanged = c.SearchSuggestions

	c.EntryHeight = -1

	c.PopupPosition = fyne.NewPos(-1, -1)

	c.itemHeight = 30

	c.Options = options

	c.Suggestions = options

	c.navigableList = newNavigableList(
		c.setTextFromMenu,
		c.HideCompletion,
		c.CustomCreate,
		c.CustomUpdate,
	)

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		c.HideCompletion()
	})

	closeBtn.Resize(fyne.NewSize(32, 32))
	closeBtn.Move(fyne.NewPos(0, 0))

	c.container = container.NewStack(
		canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground)),
		container.New(
			&CompletionListEntryLayout{},
			c.navigableList,
			closeBtn,
		),
	)

	// Caching the suggestions, this is needed for faster search!
	go func() {
		c.lowerSuggestions = make([]string, len(c.Suggestions))
		for i, s := range c.Suggestions {
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
		// JC.Logln("Skipping duplicate input:", s)
		return
	}
	c.lastInput = s

	if c.pause {
		JC.MainDebouncer.Cancel("show_suggestion")
		// JC.Logln("Cancelling due to pause?")
		return
	}

	delay := 50 * time.Millisecond
	if c.popup.Visible() {
		delay = 100 * time.Millisecond

	}
	if JC.IsMobile {
		if c.popup.Visible() {
			delay = 200 * time.Millisecond
		}
	}

	minText := 1

	// Bail out early
	if len(s) < minText || s == "" {
		JC.MainDebouncer.Cancel("show_suggestion")
		fyne.Do(func() {
			c.HideCompletion()
		})
		return
	}

	// JC.Logln("Registering debouncer call for", s)

	JC.MainDebouncer.Call("show_suggestion", delay, func() {

		input := c.GetCurrentInput()

		// JC.Logln("Debounced trigger for:", input)

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
				display := c.Suggestions[i]
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

		if JC.EqualStringSlices(results, c.Options) {
			// JC.Logln("Same results, skipping UI update")
			return
		}

		fyne.Do(func() {
			c.SetOptions(results)
			c.ShowCompletion()
		})
	})
}

func (c *CompletionEntry) TypedKey(event *fyne.KeyEvent) {
	// Fyne weird. without this backspace doesnt work?
	c.Entry.TypedKey(event)

	// Seems redundant?
	// c.SearchSuggestions(c.Text)
}

func (c *CompletionEntry) FocusLost() {

	c.Entry.FocusLost()

	if JC.IsMobile {

		// Fix for when android keyboard hiding, position got bad
		JC.MainDebouncer.Call("completion_entry_positioning", 100*time.Millisecond, func() {
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
		c.navigableList.SetFilteredData(c.Options)
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
}

func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
}

func (c *CompletionEntry) Resize(size fyne.Size) {
	c.Entry.Resize(size)

	if c.popup != nil {
		c.popup.Resize(c.maxSize())
		c.popup.Move(c.popUpPos())
	}
}

func (c *CompletionEntry) SetOptions(itemList []string) {

	c.Options = itemList

	if c.navigableList != nil {
		c.navigableList.SetFilteredData(c.Options)
	}
}

func (c *CompletionEntry) ShowCompletion() {
	if c.pause {
		// JC.Logln("Entry is paused")
		return
	}

	if len(c.Options) == 0 || len(c.Text) == 0 {
		// JC.Logln("Entry has no options")
		c.HideCompletion()
		return
	}

	if c.popup.Visible() && len(c.popup.Objects) != 0 {
		// JC.Logln("Popup already visible, not recalculating position again")
		return
	}

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

	if refresh {
		canvas.Refresh(c.popup)
	}

	if len(c.popup.Objects) == 0 {
		c.popup.Add(c.container)
	}

	c.popup.Show()

	ActiveEntry = c
}

func (c *CompletionEntry) calculatePosition() bool {

	if c.Canvas == nil {
		c.Canvas = fyne.CurrentApp().Driver().CanvasForObject(c)
	}

	if c.Canvas != nil {
		c.Scale = c.Canvas.Scale()
	}

	if c.Canvas == nil || c.Scale == 0 {
		return false
	}

	if c.Parent == nil || c.Parent.overlayContent == nil {
		return false
	}

	p := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	x := fyne.CurrentApp().Driver().AbsolutePositionForObject(c.Parent.overlayContent)
	px := p.Subtract(x)

	c.PopupPosition = px

	c.EntryHeight = c.Size().Height

	c.EntryWidth = c.Size().Width

	return true
}

func (c *CompletionEntry) maxSize() fyne.Size {

	if !c.calculatePosition() {
		return fyne.NewSize(0, 0)
	}

	if c.Canvas == nil {
		return fyne.NewSize(0, 0)
	}

	// Disabling Dynamic height for now.
	// padding := (theme.Padding() * 2) * c.Scale
	// separator := theme.SeparatorThicknessSize()

	// listHeight := float32(len(c.Options))*(c.itemHeight+padding+separator) + padding
	// maxHeight := c.Canvas.Size().Height - c.PopupPosition.Y - c.EntryHeight - padding

	// if maxHeight > 300 {
	// 	maxHeight = 300
	// }

	// if JC.IsMobile {
	// 	maxHeight = 200
	// }

	// if listHeight > maxHeight {
	// 	listHeight = maxHeight
	// }

	listHeight := float32(300)
	if JC.IsMobile {
		listHeight = float32(200)
	}

	return fyne.NewSize(c.EntryWidth, listHeight)
}

func (c *CompletionEntry) popUpPos() fyne.Position {
	if !c.calculatePosition() {
		return fyne.NewPos(0, 0)
	}

	entryPos := c.PopupPosition
	entryPos.Y += c.EntryHeight

	return entryPos

}

func (c *CompletionEntry) setTextFromMenu(s string) {
	JC.MainDebouncer.Cancel("show_suggestion")
	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.popup.Hide()
	c.pause = false
}

type navigableList struct {
	widget.List
	selected        int
	setTextFromMenu func(string)
	hide            func()
	filteredData    []string
	navigating      bool
	customCreate    func() fyne.CanvasObject
	customUpdate    func(id widget.ListItemID, object fyne.CanvasObject)
}

func newNavigableList(
	setTextFromMenu func(string),
	hide func(),
	create func() fyne.CanvasObject,
	update func(id widget.ListItemID, object fyne.CanvasObject),
) *navigableList {

	n := &navigableList{
		selected:        -1,
		setTextFromMenu: setTextFromMenu,
		hide:            hide,
		customCreate:    create,
		customUpdate:    update,
	}

	visibleCount := 10

	if JC.IsMobile {
		visibleCount = 6
	}

	n.List = widget.List{
		Length: func() int {
			if visibleCount > len(n.filteredData) {
				return len(n.filteredData)
			}
			return visibleCount
		},
		CreateItem: func() fyne.CanvasObject {
			if fn := n.customCreate; fn != nil {
				return fn()
			}
			return NewSelectableText()
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			// Lazy reveal logic
			if i >= visibleCount-10 && visibleCount < len(n.filteredData) {
				// oldCount := visibleCount
				visibleCount += 50
				if visibleCount > len(n.filteredData) {
					visibleCount = len(n.filteredData)
				}
				// JC.Logln("Lazy refresh: showing", visibleCount, "of", len(n.filteredData), "items (added", visibleCount-oldCount, ")")
				n.Refresh()
			}

			// Update item content
			item := n.filteredData[i]
			if st, ok := o.(*SelectableText); ok {
				st.SetText(item)
				st.SetIndex(i)
				st.SetParent(n)
			}
		},
		OnSelected: func(i widget.ListItemID) {
			if !n.navigating && i > -1 {
				item := n.filteredData[i]
				n.setTextFromMenu(item)
			}
			n.navigating = false
		},
	}

	n.ExtendBaseWidget(n)

	return n
}
func (n *navigableList) FocusGained() {
}

func (n *navigableList) FocusLost() {
}

func (n *navigableList) SetFilteredData(items []string) {
	if JC.EqualStringSlices(n.filteredData, items) {
		// JC.Logln("Same filtered data, skipping update")
		return
	}
	n.Unselect(n.selected)
	n.filteredData = items
	n.Refresh()
	n.selected = -1

	// JC.Logln("Injecting filtered view with", len(items), "items")
}

func (n *navigableList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.filteredData)-1 {
			n.selected++
		} else {
			n.selected = 0
		}
		n.navigating = true
		n.Select(n.selected)

	case fyne.KeyUp:
		if n.selected > 0 {
			n.selected--
		} else {
			n.selected = len(n.filteredData) - 1
		}
		n.navigating = true
		n.Select(n.selected)
	case fyne.KeyEscape:
		n.hide()

	}
}

func (n *navigableList) TypedRune(r rune) {}

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

	// Layout close button on the right
	closeBtn.Resize(fyne.NewSize(closeWidth, height))
	closeBtn.Move(fyne.NewPos(size.Width-closeWidth, 0))

	// Layout list entry to fill remaining space
	listEntry.Resize(fyne.NewSize(size.Width-closeWidth, height))
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
