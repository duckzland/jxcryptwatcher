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

var activeEntry *completionEntry = nil

type completionEntry struct {
	widget.Entry
	popup          *fyne.Container
	container      *fyne.Container
	completionList *completionList
	parent         DialogForm
	pause          bool
	shifted        bool
	itemHeight     float32
	shiftX         float32
	options        []string
	popupPosition  fyne.Position
	entryPosition  fyne.Position
	canvas         fyne.Canvas
	lastInput      string
	action         func(active bool)
	worker         *completionWorker
}

func NewCompletionEntry(
	options []string,
	searchOptions []string,
	popup *fyne.Container,
) *completionEntry {

	delay := 16 * time.Millisecond
	if JC.IsMobile {
		delay = 50 * time.Millisecond
	}

	c := &completionEntry{
		options:       options,
		popup:         popup,
		popupPosition: fyne.NewPos(-1, -1),
		entryPosition: fyne.NewPos(-1, -1),
		itemHeight:    40,
		shifted:       false,
		shiftX:        36,
		worker: &completionWorker{
			searchable: searchOptions,
			data:       options,
			total:      len(searchOptions),
			chunk:      len(searchOptions) / JC.MaximumThreads(4),
			delay:      delay,
		},
	}

	c.ExtendBaseWidget(c)

	c.OnChanged = c.searchSuggestions

	c.worker.Init()

	c.completionList = NewCompletionList(c.setTextFromMenu, c.hideCompletion, c.itemHeight)

	cLayout := &completionListEntryLayout{
		background: canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground)),
		listEntry:  c.completionList,
		closeSize:  fyne.NewSize(32, 32),
		closeBtn: NewActionButton("close_entry", "", theme.CancelIcon(), "", "normal", func(btn ActionButton) {
			c.hideCompletion()
		}, nil),
	}

	c.container = container.New(
		cLayout,
		cLayout.background,
		cLayout.listEntry,
		cLayout.closeBtn,
	)

	c.popup.Add(c.container)

	c.popup.Hide()

	return c
}

func (c *completionEntry) TypedKey(event *fyne.KeyEvent) {
	c.Entry.TypedKey(event)
}

func (c *completionEntry) FocusLost() {
	c.Entry.FocusLost()

	if c.action != nil {
		c.action(false)
	}
}

func (c *completionEntry) FocusGained() {

	c.Entry.FocusGained()

	if activeEntry != nil && activeEntry != c {
		activeEntry.hideCompletion()
	}

	if len(c.Text) > 0 {
		c.searchSuggestions(c.Text)
		c.completionList.SetData(c.options)
		c.showCompletion()
	}

	if c.action != nil {
		c.action(true)
	}
}

func (c *completionEntry) Refresh() {
	c.Entry.Refresh()
}

func (c *completionEntry) Resize(size fyne.Size) {

	if c.shifted {
		size.Width -= c.shiftX
	}

	if c.Entry.Size() != size {
		c.Entry.Resize(size)
	}

	if c.popup != nil && c.popup.Visible() {
		c.shiftEntry()

		c.popupPosition = fyne.NewPos(-1, -1)
		np := c.popUpPos()

		if c.popup.Position() != np {
			c.popup.Move(np)
		}

		c.dynamicResize()

	}
}

func (c *completionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *completionEntry) SetAction(fn func(active bool)) {
	c.action = fn
}

func (c *completionEntry) SetValidator(fn func(string) error) {
	c.Validator = fn
}

func (c *completionEntry) SetParent(parent DialogForm) {
	c.parent = parent
}

func (c *completionEntry) setOptions(itemList []string) {
	c.options = itemList
	if c.completionList != nil {
		c.completionList.SetData(c.options)
	}
}

func (c *completionEntry) shiftEntry() {
	if c.popup != nil && c.popup.Visible() && !c.shifted {
		cs := c.Size()
		cs.Width -= c.shiftX
		if c.Entry.Size() != cs {
			c.Entry.Resize(cs)
		}

		c.shifted = true
	}

}

func (c *completionEntry) unshiftEntry() {
	if c.popup != nil && !c.popup.Visible() && c.shifted {
		cs := c.Size()
		cs.Width += c.shiftX

		if c.Entry.Size() != cs {
			c.Entry.Resize(cs)
		}

		c.shifted = false
	}
}

func (c *completionEntry) dynamicResize() {
	mx := c.maxSize()
	ox := c.popup.Size()

	if mx != ox {
		c.popup.Resize(mx)
	}
}

func (c *completionEntry) hideCompletion() {

	if c.popup != nil {
		c.popup.Hide()
	}

	c.popupPosition = fyne.NewPos(-1, -1)

	c.unshiftEntry()
}

func (c *completionEntry) showCompletion() {

	if c.pause {
		return
	}

	if len(c.options) == 0 || len(c.Text) == 0 {
		c.hideCompletion()
		return
	}

	if c.popup.Visible() && c.popupPosition.Y != -1 {
		return
	}

	// Always reset position cache!
	c.calculatePosition(true)

	mx := c.maxSize()
	ox := c.popup.Size()

	mp := c.popUpPos()
	op := c.popup.Position()

	if mx.Width != ox.Width {
		c.popup.Resize(mx)
	}

	if mp != op {
		c.popup.Move(mp)
	}

	c.popup.Show()

	c.shiftEntry()

	activeEntry = c
}

func (c *completionEntry) getCurrentInput() string {
	return c.lastInput
}

func (c *completionEntry) searchSuggestions(s string) {

	c.worker.Cancel()

	if c.pause {
		return
	}

	if len(s) < 1 {
		c.hideCompletion()
		return
	}

	c.lastInput = strings.ToLower(s)

	c.worker.Search(s, func(input string, results []string) {
		if len(results) == 0 {
			fyne.Do(func() {
				c.hideCompletion()
			})
			return
		}

		if input != c.getCurrentInput() {
			return
		}

		fyne.Do(func() {
			if input == c.getCurrentInput() {

				JC.Logln("Performing UX Update", input, len(results))

				c.setOptions(results)
				c.showCompletion()
				c.dynamicResize()
			}
		})
	})
}

func (c *completionEntry) calculatePosition(force bool) bool {

	if c.canvas == nil {
		c.canvas = fyne.CurrentApp().Driver().CanvasForObject(c)
	}

	if c.canvas == nil {
		return false
	}

	if c.parent == nil || c.parent.GetContent() == nil {
		return false
	}

	if c.popupPosition.Y != -1 && !force {
		return true
	}

	p := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	x := fyne.CurrentApp().Driver().AbsolutePositionForObject(c.parent.GetContent())

	c.entryPosition = fyne.CurrentApp().Driver().AbsolutePositionForObject(c.parent.GetForm())
	c.popupPosition = p.Subtract(x)

	return true
}

func (c *completionEntry) maxSize() fyne.Size {

	if !c.calculatePosition(false) {
		return fyne.NewSize(0, 0)
	}

	if c.canvas == nil {
		return fyne.NewSize(0, 0)
	}

	listHeight := float32(len(c.options)) * c.itemHeight
	maxHeight := c.canvas.Size().Height - c.popupPosition.Y - c.Size().Height

	if maxHeight > 300 {
		maxHeight = 7 * c.itemHeight
	}

	if listHeight > maxHeight {
		listHeight = maxHeight
	}

	width := c.Size().Width

	if c.shifted {
		width += c.shiftX
	}

	return fyne.NewSize(width, listHeight)
}

func (c *completionEntry) popUpPos() fyne.Position {
	if !c.calculatePosition(false) {
		return fyne.NewPos(0, 0)
	}

	entryPos := c.popupPosition
	entryPos.Y += c.Size().Height
	entryPos.Y += 2

	return entryPos

}

func (c *completionEntry) setTextFromMenu(s string) {
	c.worker.Cancel()

	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.popup.Hide()
	c.unshiftEntry()
	c.pause = false
}
