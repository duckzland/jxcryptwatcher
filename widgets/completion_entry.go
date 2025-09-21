package widgets

import (
	"context"
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

var ActiveEntry *completionEntry = nil

type completionEntry struct {
	widget.Entry
	popup               *fyne.Container
	container           *fyne.Container
	completionList      *completionList
	parent              *DialogForm
	pause               bool
	itemHeight          float32
	options             []string
	suggestions         []string
	searchable          []string
	searchableTotal     int
	searchableChunkSize int
	searchCancel        context.CancelFunc
	popupPosition       fyne.Position
	canvas              fyne.Canvas
	lastInput           string
	uuid                string
	userValidator       func(string) error
	skipValidation      bool
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
		searchableChunkSize: min((len(searchOptions)+JC.HWTotalCPU-1)/JC.HWTotalCPU, 4),
		popup:               popup,
		popupPosition:       fyne.NewPos(-1, -1),
		itemHeight:          40,
	}
	c.ExtendBaseWidget(c)

	c.OnChanged = func(s string) {
		c.SearchSuggestions(s)
		JC.MainDebouncer.Call("Validating-"+c.uuid, 300*time.Millisecond, func() {
			fyne.Do(func() {
				if !c.popup.Visible() {
					c.skipValidation = false
					c.SetValidationError(c.Validator(c.Text))
				}
			})
		})
	}

	c.completionList = NewCompletionList(c.setTextFromMenu, c.HideCompletion, c.itemHeight)

	closeBtn := NewActionButton("close_entry", "", theme.CancelIcon(), "", "normal", func(btn *ActionButton) {
		c.HideCompletion()
	}, nil)

	closeBtn.Resize(fyne.NewSize(32, 32))
	closeBtn.Move(fyne.NewPos(0, 0))

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	bg.CornerRadius = JC.PanelBorderRadius

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

func (c *completionEntry) GetCurrentInput() string {
	return c.lastInput
}

func (c *completionEntry) SearchSuggestions(s string) {

	if s == c.lastInput {
		return
	}

	c.lastInput = s

	if c.pause {
		c.CancelSearch()
		return
	}

	delay := 10 * time.Millisecond
	if JC.IsMobile || c.popup.Visible() {
		delay = 50 * time.Millisecond
	}

	minText := 1

	c.CancelSearch()

	if len(s) < minText || s == "" {
		fyne.Do(func() {
			c.HideCompletion()
		})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.searchCancel = cancel

	JC.MainDebouncer.Call("show_suggestion_"+c.uuid, delay, func() {
		input := c.GetCurrentInput()

		if len(input) < minText || input == "" {
			cancel()
			fyne.Do(func() {
				c.HideCompletion()
			})
			return
		}

		lowerS := strings.ToLower(input)

		var wg sync.WaitGroup
		var mu sync.Mutex
		results := []string{}

		for i := 0; i < len(c.searchable); i += c.searchableChunkSize {
			end := min(i+c.searchableChunkSize, c.searchableTotal)

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				local := []string{}
				for j := start; j < end; j++ {
					if ctx.Err() != nil {
						return
					}
					if strings.Contains(c.searchable[j], lowerS) {
						local = append(local, c.suggestions[j])
					}
				}
				if ctx.Err() != nil {
					return
				}
				mu.Lock()
				results = append(results, local...)
				mu.Unlock()
			}(i, end)
		}

		wg.Wait()

		if ctx.Err() != nil {
			return
		}

		if len(results) == 0 {
			fyne.Do(func() {
				c.HideCompletion()
			})
			return
		}

		if input != c.GetCurrentInput() {
			return
		}

		results = JC.ReorderSearchable(results)

		fyne.Do(func() {
			c.SetOptions(results)
			c.ShowCompletion()
			c.DynamicResize()
		})
	})
}

func (c *completionEntry) CancelSearch() {
	JC.MainDebouncer.Cancel("show_suggestion_" + c.uuid)
	if c.searchCancel != nil {
		c.searchCancel()
		c.searchCancel = nil
	}
}

func (c *completionEntry) TypedKey(event *fyne.KeyEvent) {
	c.Entry.TypedKey(event)
	c.skipValidation = true
	JC.MainDebouncer.Cancel("validating-" + c.uuid)
}

func (c *completionEntry) FocusLost() {
	c.Entry.FocusLost()
}

func (c *completionEntry) FocusGained() {

	c.SetValidationError(nil)

	c.Entry.FocusGained()

	if ActiveEntry != nil && ActiveEntry != c {
		ActiveEntry.HideCompletion()
	}

	if len(c.Text) > 0 {
		c.completionList.SetData(c.options)
		c.ShowCompletion()
	}
}

func (c *completionEntry) SetDefaultValue(s string) {
	c.Text = s
}

func (c *completionEntry) HideCompletion() {

	c.skipValidation = false

	if c.popup != nil {
		c.popup.Hide()
	}

	if c.completionList != nil {
		c.completionList.SetData([]string{})
	}

	c.popupPosition = fyne.NewPos(-1, -1)
}

func (c *completionEntry) Refresh() {
	c.Entry.Refresh()
}

func (c *completionEntry) Resize(size fyne.Size) {

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

func (c *completionEntry) SetValidator(fn func(string) error) {
	c.Validator = func(s string) error {
		if c.skipValidation {
			c.SetValidationError(nil)
			return nil
		}

		return fn(s)
	}
}

func (c *completionEntry) Validate() error {
	// This is important to prevent user trying to click save when skipvalidation is active!
	c.skipValidation = false
	err := c.Validator(c.Entry.Text)
	c.SetValidationError(err)

	return err
}

func (c *completionEntry) DynamicResize() {
	mx := c.maxSize()
	ox := c.popup.Size()

	if mx.Width != ox.Width || mx.Height != ox.Height {
		c.popup.Resize(mx)
		canvas.Refresh(c.popup)
	}
}

func (c *completionEntry) SetOptions(itemList []string) {

	c.options = itemList

	if c.completionList != nil {
		c.completionList.SetData(c.options)
	}
}

func (c *completionEntry) SetParent(parent *DialogForm) {
	c.parent = parent
}

func (c *completionEntry) ShowCompletion() {

	c.skipValidation = true

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

	c.completionList.selected = -1

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

func (c *completionEntry) calculatePosition(force bool) bool {

	if c.canvas == nil {
		c.canvas = fyne.CurrentApp().Driver().CanvasForObject(c)
	}

	if c.canvas == nil {
		return false
	}

	if c.parent == nil || c.parent.content == nil {
		return false
	}

	if c.popupPosition.Y != -1 && !force {
		return true
	}

	p := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	x := fyne.CurrentApp().Driver().AbsolutePositionForObject(c.parent.content)
	px := p.Subtract(x)

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

	listHeight := float32(len(c.options)) * c.itemHeight
	maxHeight := c.canvas.Size().Height - c.popupPosition.Y - c.Size().Height

	if maxHeight > 300 {
		maxHeight = 7 * c.itemHeight
	}

	if listHeight > maxHeight {
		listHeight = maxHeight
	}

	return fyne.NewSize(c.Size().Width, listHeight)
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
	c.CancelSearch()

	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.popup.Hide()
	c.pause = false

	c.skipValidation = false
	c.SetValidationError(c.Validator(c.Text))

}
