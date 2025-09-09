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
	popup         *fyne.Container
	container     *fyne.Container
	navigableList *navigableList
	pause         bool
	itemHeight    float32
	Parent        *ExtendedFormDialog
	Options       []string
	Suggestions   []string
	CustomCreate  func() fyne.CanvasObject
	CustomUpdate  func(id widget.ListItemID, object fyne.CanvasObject)
	PopupPosition fyne.Position
	EntryHeight   float32
	EntryWidth    float32
	Canvas        fyne.Canvas
	Scale         float32
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

	c.SetOptions(options)

	return c
}

func (c *CompletionEntry) SearchSuggestions(s string) {

	if c.pause {
		JC.MainDebouncer.Cancel("show_suggestion")
		return
	}

	minText := 1

	// completion start for text length >= 3
	if len(s) < minText {
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

	delay := 50 * time.Millisecond

	// Mobile uses virtual keyboard, give more time for user to type
	if JC.IsMobile {
		delay = 100 * time.Millisecond
	}

	// then show them
	JC.MainDebouncer.Call("show_suggestion", delay, func() {
		fyne.Do(func() {
			c.SetOptions(results)
			c.ShowCompletion()
		})
	})
}

func (c *CompletionEntry) TypedKey(event *fyne.KeyEvent) {
	c.Entry.TypedKey(event)
	c.SearchSuggestions(c.Text)
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
		c.ShowCompletion()
	}
}

func (c *CompletionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *CompletionEntry) HideCompletion() {
	if c.popup != nil {
		c.popup.Objects = nil
		c.popup.Hide()
		c.popup.Refresh()
	}

	JC.MainDebouncer.Cancel("show_suggestion")
}

func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
	if c.navigableList != nil {
		c.navigableList.SetOptions(c.Options)
	}
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
	c.Refresh()
}

func (c *CompletionEntry) CreateList() {
	if c.navigableList == nil {
		c.navigableList = newNavigableList(
			c.Options,
			c.setTextFromMenu,
			c.HideCompletion,
			c.CustomCreate,
			c.CustomUpdate,
		)
	}

	if c.container == nil {
		closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
			c.HideCompletion()
		})

		cc := container.New(
			&CompletionListEntryLayout{},
			c.navigableList,
			closeBtn,
		)

		bg := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))

		c.container = container.NewStack(bg, cc)

		closeBtn.Resize(fyne.NewSize(32, 32))
		closeBtn.Move(fyne.NewPos(0, 0))
	}

	if len(c.popup.Objects) == 0 {
		c.popup.Add(c.container)
	}
}

func (c *CompletionEntry) ShowCompletion() {

	JC.MainDebouncer.Cancel("show_suggestion")

	if c.pause {
		return
	}

	if len(c.Options) == 0 {
		c.HideCompletion()
		return
	}

	c.CreateList()

	c.navigableList.UnselectAll()
	c.navigableList.selected = -1

	mx := c.maxSize()

	c.popup.Resize(mx)
	c.popup.Move(c.popUpPos())

	canvas.Refresh(c.popup)

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
	// c.EntryHeight += (theme.Padding() * 2) * c.Scale

	// if JC.IsMobile {
	// 	c.EntryHeight += (theme.InputBorderSize() * 2) * c.Scale

	// 	// Hackish, different device, different android version have different height..
	// 	// Not sure how to properly get precise height value across different device and android version yet.
	// 	c.EntryHeight += 8 * c.Scale
	// }

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

	padding := (theme.Padding() * 2) * c.Scale
	separator := theme.SeparatorThicknessSize()

	listHeight := float32(len(c.Options))*(c.itemHeight+padding+separator) + padding
	maxHeight := c.Canvas.Size().Height - c.PopupPosition.Y - c.EntryHeight - padding

	if maxHeight > 300 {
		maxHeight = 300
	}

	if JC.IsMobile {
		maxHeight = 200
	}
	if listHeight > maxHeight {
		listHeight = maxHeight
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
	navigating      bool
	items           []string

	customCreate func() fyne.CanvasObject
	customUpdate func(id widget.ListItemID, object fyne.CanvasObject)
}

func newNavigableList(
	items []string,
	setTextFromMenu func(string),
	hide func(),
	create func() fyne.CanvasObject,
	update func(id widget.ListItemID, object fyne.CanvasObject),
) *navigableList {

	n := &navigableList{
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
	case fyne.KeyEscape:
		n.hide()

	}
}

func (n *navigableList) TypedRune(r rune) {}

type CompletionListEntryLayout struct{}

func (l *CompletionListEntryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

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

	width := listMin.Width + closeMin.Width // fixed close button width
	height := fyne.Max(listMin.Height, closeMin.Height)

	return fyne.NewSize(width, height)
}
