package widgets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type CompletionEntry struct {
	widget.Entry
	popupMenu     *widget.PopUp
	navigableList *navigableList
	Options       []string
	pause         bool
	itemHeight    float32
	Suggestions   []string
	CustomCreate  func() fyne.CanvasObject
	CustomUpdate  func(id widget.ListItemID, object fyne.CanvasObject)
}

func NewCompletionEntry(
	options []string,
) *CompletionEntry {

	c := &CompletionEntry{Suggestions: options}
	c.ExtendBaseWidget(c)

	if !JC.IsMobile {
		c.OnChanged = c.SearchSuggestions
	}

	c.SetOptions(options)

	return c
}

func (c *CompletionEntry) SearchSuggestions(s string) {

	// completion start for text length >= 3
	if len(s) < 1 {
		c.HideCompletion()
		return
	}

	results := []string{}
	// Search the text
	for _, part := range c.Suggestions {
		if strings.Contains(strings.ToLower(part), strings.ToLower(s)) {
			results = append(results, part)
		}
	}

	// no results
	if len(results) == 0 {
		c.HideCompletion()
		return
	}

	results = JC.ReorderByMatch(results, s)

	// then show them
	c.SetOptions(results)
	c.ShowCompletion()
}

func (c *CompletionEntry) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		if JC.IsMobile {
			c.SearchSuggestions(c.Text)
		}
	}

	c.Entry.TypedKey(event)
}

func (c *CompletionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *CompletionEntry) HideCompletion() {
	if c.popupMenu != nil {
		c.popupMenu.Hide()
	}
}

func (c *CompletionEntry) Move(pos fyne.Position) {
	c.Entry.Move(pos)
	if c.popupMenu != nil {
		c.popupMenu.Resize(c.maxSize())
		c.popupMenu.Move(c.popUpPos())
	}
}

func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
	if c.navigableList != nil {
		c.navigableList.SetOptions(c.Options)
	}
}

func (c *CompletionEntry) Resize(size fyne.Size) {
	c.Entry.Resize(size)
	if c.popupMenu != nil {
		c.popupMenu.Resize(c.maxSize())
	}
}

func (c *CompletionEntry) SetOptions(itemList []string) {
	c.Options = itemList
	c.Refresh()
}

func (c *CompletionEntry) ShowCompletion() {
	if c.pause {
		return
	}
	if len(c.Options) == 0 {
		c.HideCompletion()
		return
	}

	if c.navigableList == nil {
		c.navigableList = newNavigableList(
			c.Options,
			&c.Entry,
			c.setTextFromMenu,
			c.HideCompletion,
			c.CustomCreate,
			c.CustomUpdate,
		)
	} else {
		c.navigableList.UnselectAll()
		c.navigableList.selected = -1
	}
	holder := fyne.CurrentApp().Driver().CanvasForObject(c)

	if c.popupMenu == nil {
		c.popupMenu = widget.NewPopUp(c.navigableList, holder)
	}
	c.popupMenu.Resize(c.maxSize())
	c.popupMenu.ShowAtPosition(c.popUpPos())
	holder.Focus(c.navigableList)
}

func (c *CompletionEntry) maxSize() fyne.Size {
	cnv := fyne.CurrentApp().Driver().CanvasForObject(c)

	if cnv == nil {
		return fyne.NewSize(0, 0)
	}

	if c.itemHeight == 0 {
		c.itemHeight = c.navigableList.CreateItem().MinSize().Height
	}

	canvasSize := cnv.Size()
	entrySize := c.Size()
	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	listHeight := float32(len(c.Options))*(c.itemHeight+2*theme.Padding()+theme.SeparatorThicknessSize()) + 2*theme.Padding()
	maxHeight := canvasSize.Height - entryPos.Y - entrySize.Height - 2*theme.Padding()

	if listHeight > maxHeight {
		listHeight = maxHeight
	}

	return fyne.NewSize(entrySize.Width, listHeight)
}

func (c *CompletionEntry) popUpPos() fyne.Position {
	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	return entryPos.Add(fyne.NewPos(0, c.Size().Height))
}

func (c *CompletionEntry) setTextFromMenu(s string) {
	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.pause = false
	c.popupMenu.Hide()
}

type navigableList struct {
	widget.List
	entry           *widget.Entry
	selected        int
	setTextFromMenu func(string)
	hide            func()
	navigating      bool
	items           []string

	customCreate func() fyne.CanvasObject
	customUpdate func(id widget.ListItemID, object fyne.CanvasObject)
}

func newNavigableList(
	items []string,
	entry *widget.Entry,
	setTextFromMenu func(string),
	hide func(),
	create func() fyne.CanvasObject,
	update func(id widget.ListItemID, object fyne.CanvasObject),
) *navigableList {

	n := &navigableList{
		entry:           entry,
		selected:        -1,
		setTextFromMenu: setTextFromMenu,
		hide:            hide,
		items:           items,
		customCreate:    create,
		customUpdate:    update,
	}

	n.List = widget.List{
		Length: func() int {
			return len(n.items)
		},
		CreateItem: func() fyne.CanvasObject {
			if fn := n.customCreate; fn != nil {
				return fn()
			}
			return widget.NewLabel("")
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			if fn := n.customUpdate; fn != nil {
				fn(i, o)
				return
			}
			o.(*widget.Label).SetText(n.items[i])
		},
		OnSelected: func(id widget.ListItemID) {
			if !n.navigating && id > -1 {
				setTextFromMenu(n.items[id])
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

func (n *navigableList) SetOptions(items []string) {
	n.Unselect(n.selected)
	n.items = items
	n.Refresh()
	n.selected = -1
}

func (n *navigableList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.items)-1 {
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
			n.selected = len(n.items) - 1
		}
		n.navigating = true
		n.Select(n.selected)
	case fyne.KeyReturn, fyne.KeyEnter:
		if n.selected == -1 {
			n.hide()
			n.entry.TypedKey(event)
		} else {
			n.navigating = false
			n.OnSelected(n.selected)
		}
	case fyne.KeyEscape:
		n.hide()
	default:
		n.entry.TypedKey(event)

	}
}

func (n *navigableList) TypedRune(r rune) {
	n.entry.TypedRune(r)
}
