package widgets

import (
	"strings"
	"sync"
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
	popup               *fyne.Container
	container           *fyne.Container
	completionList      *completionList
	parent              DialogForm
	pause               bool
	itemHeight          float32
	options             []string
	suggestions         []string
	searchable          []string
	searchableTotal     int
	searchableChunkSize int
	popupPosition       fyne.Position
	entryPosition       fyne.Position
	canvas              fyne.Canvas
	lastInput           string
	uuid                string
	dispatcher          JC.Dispatcher
	action              func(active bool)
}

func NewCompletionEntry(
	options []string,
	searchOptions []string,
	popup *fyne.Container,
) *completionEntry {

	c := &completionEntry{
		uuid:                JC.CreateUUID(),
		options:             options,
		suggestions:         options,
		searchable:          searchOptions,
		searchableTotal:     len(searchOptions),
		searchableChunkSize: (len(searchOptions) + JC.TotalCPU() - 1) / JC.TotalCPU(),
		popup:               popup,
		popupPosition:       fyne.NewPos(-1, -1),
		entryPosition:       fyne.NewPos(-1, -1),
		itemHeight:          40,
	}

	c.ExtendBaseWidget(c)

	c.OnChanged = c.searchSuggestions

	c.dispatcher = JC.NewDispatcher(1000, c.searchableChunkSize, 16*time.Millisecond)
	c.dispatcher.Start()

	c.completionList = NewCompletionList(c.setTextFromMenu, c.hideCompletion, c.itemHeight)

	closeBtn := NewActionButton("close_entry", "", theme.CancelIcon(), "", "normal", func(btn ActionButton) {
		c.hideCompletion()
	}, nil)

	closeBtn.Resize(fyne.NewSize(32, 32))
	closeBtn.Move(fyne.NewPos(0, 0))

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	bg.CornerRadius = JC.UseTheme().Size(JC.SizePanelBorderRadius)

	c.container = container.NewStack(
		bg,
		container.New(
			&completionListEntryLayout{},
			c.completionList,
			closeBtn,
		),
	)

	c.popup.Add(c.container)

	c.popup.Hide()

	return c
}

func (c *completionEntry) SetAction(fn func(active bool)) {
	c.action = fn
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
		c.completionList.SetData(c.options)
		c.showCompletion()
	}

	if c.action != nil {
		c.action(true)
	}
}

func (c *completionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *completionEntry) Refresh() {
	c.Entry.Refresh()
}

func (c *completionEntry) Resize(size fyne.Size) {
	c.Entry.Resize(size)
	if c.popup != nil && c.popup.Visible() {
		c.popupPosition = fyne.NewPos(-1, -1)
		c.popup.Move(c.popUpPos())
		c.dynamicResize()
	}
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

func (c *completionEntry) hideCompletion() {

	if c.popup != nil {
		c.popup.Hide()
	}

	if c.completionList != nil {
		c.completionList.SetData([]string{})
	}

	c.popupPosition = fyne.NewPos(-1, -1)
}

func (c *completionEntry) dynamicResize() {
	mx := c.maxSize()
	ox := c.popup.Size()

	if mx.Width != ox.Width || mx.Height != ox.Height {
		c.popup.Resize(mx)
		canvas.Refresh(c.popup)
	}
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

	activeEntry = c
}

func (c *completionEntry) getCurrentInput() string {
	return c.lastInput
}

func (c *completionEntry) searchSuggestions(s string) {

	c.dispatcher.Drain()
	c.dispatcher.Pause()

	if s == c.lastInput || c.pause {
		return
	}

	if len(s) < 1 {
		c.hideCompletion()
		return
	}

	results := []string{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	lowerInput := strings.ToLower(s)
	c.lastInput = s
	input := s
	delay := 5 * time.Millisecond
	if c.popup.Visible() {
		delay = 10 * time.Millisecond
	}

	for i := 0; i < c.searchableTotal; i += c.searchableChunkSize {
		start := i
		end := min(i+c.searchableChunkSize, c.searchableTotal)

		wg.Add(1)
		c.dispatcher.Submit(func() {
			defer wg.Done()

			local := []string{}
			for j := start; j < end; j++ {
				if strings.Contains(c.searchable[j], lowerInput) {
					local = append(local, c.suggestions[j])
				}
			}

			mu.Lock()
			results = append(results, local...)
			mu.Unlock()
		})
	}

	c.dispatcher.Resume()

	wg.Wait()

	if len(results) == 0 {
		c.hideCompletion()
		return
	}

	results = JC.ReorderSearchable(results)

	JC.UseDebouncer().Call("show_suggestion_"+c.uuid, delay, func() {
		if input != c.getCurrentInput() {
			return
		}

		fyne.Do(func() {
			c.setOptions(results)
			c.showCompletion()
			c.dynamicResize()
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
	z := fyne.CurrentApp().Driver().AbsolutePositionForObject(c.parent.GetForm())
	px := p.Subtract(x)

	c.entryPosition = z
	c.popupPosition = px

	return true
}

func (c *completionEntry) maxSize() fyne.Size {

	if !c.calculatePosition(false) {
		return fyne.NewSize(0, 0)
	}

	if c.canvas == nil {
		return fyne.NewSize(0, 0)
	}

	size := c.Size()
	listHeight := float32(len(c.options)) * c.itemHeight
	maxHeight := c.canvas.Size().Height - c.popupPosition.Y - c.Size().Height

	if maxHeight > 300 {
		maxHeight = 7 * c.itemHeight
	}

	if listHeight > maxHeight {
		listHeight = maxHeight
	}

	width := size.Width
	if size.Width < 300 {
		width = (c.popupPosition.X - c.entryPosition.X) + width - 20
	}

	return fyne.NewSize(width, listHeight)
}

func (c *completionEntry) popUpPos() fyne.Position {
	if !c.calculatePosition(false) {
		return fyne.NewPos(0, 0)
	}

	size := c.Size()
	entryPos := c.popupPosition
	entryPos.Y += size.Height
	entryPos.Y += 2

	if size.Width < 300 {
		entryPos.X = c.entryPosition.X + 20
	}

	return entryPos

}

func (c *completionEntry) setTextFromMenu(s string) {
	JC.UseDebouncer().Cancel("show_suggestion_" + c.uuid)

	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.popup.Hide()
	c.pause = false
}
