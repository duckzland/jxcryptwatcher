package widgets

import (
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type completionList struct {
	widget.BaseWidget
	selected         int
	onChange         func(string)
	onClose          func()
	filteredData     []string
	scrollOffset     int
	itemHeight       float32
	contentBox       *fyne.Container
	scrollContent    *canvas.Rectangle
	scrollBox        *container.Scroll
	layout           *completionListLayout
	root             *fyne.Container
	lastSize         fyne.Size
	fps              time.Duration
	position         fyne.Position
	cursorOffset     fyne.Position
	dragging         bool
	done             chan struct{}
	maxOffsetY       float32
	itemVisible      int
	scaledItemHeight float32
}

func NewCompletionList(
	onChange func(string),
	onClose func(),
	itemHeight float32,
) *completionList {
	n := &completionList{
		selected:      -1,
		onChange:      onChange,
		onClose:       onClose,
		itemHeight:    itemHeight,
		lastSize:      fyne.NewSize(0, 0),
		scrollContent: canvas.NewRectangle(JC.Transparent),
		contentBox:    container.New(layout.NewVBoxLayout()),
		fps:           time.Millisecond * 32,
		maxOffsetY:    -1,
	}

	n.scrollBox = container.NewVScroll(n.scrollContent)
	n.scrollBox.OnScrolled = n.scrollingContent

	n.root = container.New(
		&completionListLayout{
			itemHeight: n.itemHeight,
			parent:     n,
			lastSize:   fyne.NewSize(0, 0),
		},
		n.contentBox,
		n.scrollBox,
	)

	n.ExtendBaseWidget(n)

	return n
}

func (n *completionList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(n.root)
}

func (n *completionList) SetFilteredData(items []string) {

	if JC.EqualStringSlices(n.filteredData, items) {
		return
	}

	n.filteredData = items
	n.scrollOffset = 0
	n.selected = -1
	n.calculateDynamicItemSize()

	JC.MainDebouncer.Call("layout_update", 50*time.Millisecond, func() {
		fyne.Do(func() {
			n.scrollContent.SetMinSize(fyne.NewSize(1, n.scaledItemHeight*float32(len(n.filteredData))))
			n.computeMaxScrollOffset()
			n.refreshContent()
		})
	})
}

func (n *completionList) calculateDynamicItemSize() {
	ttl := float32(len(n.filteredData))

	n.scaledItemHeight = n.itemHeight

	oh := n.itemHeight * ttl
	mx := float32(1000)
	if oh > mx {
		n.scaledItemHeight = float32(math.Ceil(float64(mx/ttl + float32(n.itemVisible))))
	}
}

func (n *completionList) Resize(size fyne.Size) {
	n.lastSize = size
	n.BaseWidget.Resize(size)
}

func (n *completionList) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (n *completionList) OnSelected(index int) {
	if index >= 0 && index < len(n.filteredData) {
		n.onChange(n.filteredData[index])
	}
}

func (n *completionList) TypedKey(event *fyne.KeyEvent) {

	if len(n.filteredData) == 0 {
		n.selected = -1
		n.scrollOffset = 0
		n.refreshContent()
		return
	}

	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.filteredData)-1 && n.selected >= 0 {
			n.selected++
		} else {
			n.selected = 0
		}
		n.scrollOffset = n.selected
		n.refreshContent()

	case fyne.KeyUp:
		if n.selected > 0 {
			n.selected--
		} else {
			n.selected = len(n.filteredData) - 1
		}
		n.scrollOffset = n.selected
		n.refreshContent()

	case fyne.KeyEscape:
		n.onClose()
	}
}

func (n *completionList) Scrolled(ev *fyne.ScrollEvent) {
	delta := int(ev.Scrolled.DY * 3 / n.scaledItemHeight)
	n.scrollOffset -= delta
	maxOffset := int(n.scrollBox.Content.Size().Height/n.scaledItemHeight) - int(n.scrollBox.Size().Height/n.scaledItemHeight)
	if n.scrollOffset < 0 {
		n.scrollOffset = 0
	}
	if n.scrollOffset > maxOffset {
		n.scrollOffset = maxOffset
	}

	n.scrollBox.Offset = fyne.NewPos(0, float32(n.scrollOffset)*n.scaledItemHeight)

	if n.scrollBox.OnScrolled != nil {
		n.scrollBox.OnScrolled(n.scrollBox.Offset)
	}

	n.scrollBox.Refresh()
	n.refreshContent()
}

func (n *completionList) Dragged(ev *fyne.DragEvent) {
	n.position = ev.Position

	if !n.dragging {
		n.dragging = true
		n.cursorOffset = ev.Position.Subtract(n.Position())
		n.done = make(chan struct{})

		sourceY := n.Position().Y
		edgeThreshold := n.getEdgeThreshold()

		go func() {
			ticker := time.NewTicker(n.fps)
			defer ticker.Stop()

			for {
				select {
				case <-n.done:
					return
				case <-ticker.C:
					targetY := n.position.Y - n.cursorOffset.Y
					delta := targetY - sourceY

					if math.Abs(float64(delta)) < float64(n.scaledItemHeight*0.6) {
						continue
					}

					direction := 0
					if delta < -edgeThreshold {
						direction = -1
					} else if delta > edgeThreshold {
						direction = 1
					}

					if direction != 0 {
						maxDelta := n.scaledItemHeight * float32(n.itemVisible) * 0.25
						clampedDY := float32(math.Max(-float64(maxDelta), math.Min(float64(ev.Dragged.DY), float64(maxDelta))))

						scrollStep := n.getScrollStepFromDelta(clampedDY * float32(direction))

						fyne.Do(func() {
							n.scrollBy(float32(direction) * scrollStep)
						})
					}
				}
			}
		}()
	}
}

func (n *completionList) DragEnd() {
	n.dragging = false
	if n.done != nil {
		close(n.done)
		n.done = nil
	}
}

func (n *completionList) IsDragging() bool {
	return n.dragging
}

func (n *completionList) refreshContent() {
	for i, obj := range n.contentBox.Objects {
		dataIndex := n.scrollOffset + i

		item, ok := obj.(*completionText)
		if !ok {
			continue
		}
		if dataIndex >= 0 && dataIndex < len(n.filteredData) {
			item.SetText(n.filteredData[dataIndex])
			item.SetIndex(dataIndex)
		}
	}
}

func (n *completionList) getEdgeThreshold() float32 {
	viewport := n.scrollBox.Size().Height
	content := float32(len(n.filteredData)) * n.scaledItemHeight

	ratio := viewport / content
	threshold := viewport * (0.2 + ratio*0.3)

	return float32(math.Max(24, math.Min(float64(threshold), float64(viewport*0.5))))
}

func (n *completionList) getScrollStepFromDelta(deltaY float32) float32 {
	itemDelta := deltaY / n.scaledItemHeight

	maxItems := float32(math.Max(1, float64(n.itemVisible)/2))
	clamped := float32(math.Max(float64(-maxItems), math.Min(float64(itemDelta), float64(maxItems))))

	return clamped * n.scaledItemHeight
}

func (n *completionList) scrollingContent(offset fyne.Position) {
	newOffset := int(offset.Y / n.scaledItemHeight)
	maxOffset := len(n.filteredData) - len(n.contentBox.Objects)

	if newOffset > maxOffset {
		newOffset = maxOffset
	}
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset != n.scrollOffset {
		n.scrollOffset = newOffset
		n.refreshContent()
	}
}

func (n *completionList) computeMaxScrollOffset() {
	if n.scrollBox == nil || n.scrollContent == nil {
		n.maxOffsetY = 0
		return
	}

	contentHeight := n.scrollContent.MinSize().Height
	viewportHeight := n.scrollBox.Size().Height

	if contentHeight <= viewportHeight {
		n.maxOffsetY = 0
	} else {
		n.maxOffsetY = contentHeight - viewportHeight
	}
}

func (n *completionList) scrollBy(delta float32) {
	if n.scrollBox == nil {
		return
	}

	current := n.scrollBox.Offset.Y
	newOffset := current + delta

	if n.maxOffsetY == -1 {
		n.computeMaxScrollOffset()
	}

	if newOffset < 0 {
		if current > 0 {
			newOffset = 0
		} else {
			return
		}
	} else if newOffset > n.maxOffsetY {
		if current < n.maxOffsetY {
			newOffset = n.maxOffsetY
		} else {
			return
		}
	}

	n.setOffsetY(newOffset)
	n.scrollingContent(fyne.NewPos(0, newOffset))
}

func (n *completionList) setOffsetY(offset float32) {
	if n.scrollBox == nil {
		return
	}

	current := n.scrollBox.Offset.Y
	if current == offset {
		return
	}

	n.scrollBox.Offset.Y = offset
	n.scrollBox.Refresh()
}

func (n *completionList) Tapped(_ *fyne.PointEvent) {
}

func (n *completionList) TypedRune(r rune) {
}

func (n *completionList) FocusGained() {
}

func (n *completionList) FocusLost() {
}
