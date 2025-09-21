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
	uuid             string
	data             []string
	selected         int
	itemVisible      int
	itemHeight       float32
	itemTotal        int
	scaledItemHeight float32
	scaledHeight     float32
	onChange         func(string)
	onClose          func()
	contentBox       *fyne.Container
	scrollContent    *canvas.Rectangle
	scrollBox        *container.Scroll
	layout           *completionListLayout
	root             *fyne.Container
	lastSize         fyne.Size
	scrollOffset     int
	maxOffset        int
	scrollLimiter    float64
	fps              time.Duration
	position         fyne.Position
	dragging         bool
	done             chan struct{}
}

func NewCompletionList(
	onChange func(string),
	onClose func(),
	itemHeight float32,
) *completionList {
	n := &completionList{
		uuid:             JC.CreateUUID(),
		selected:         -1,
		onChange:         onChange,
		onClose:          onClose,
		itemHeight:       itemHeight,
		scaledItemHeight: itemHeight,
		scaledHeight:     0,
		itemTotal:        0,
		lastSize:         fyne.NewSize(0, 0),
		scrollContent:    canvas.NewRectangle(JC.Transparent),
		contentBox:       container.New(layout.NewVBoxLayout()),
		fps:              time.Millisecond * 32,
		maxOffset:        0,
		scrollLimiter:    0,
		scrollOffset:     0,
		dragging:         false,
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

func (n *completionList) SetData(items []string) {

	if JC.EqualStringSlices(n.data, items) {
		return
	}

	n.data = items
	n.scrollOffset = 0
	n.selected = -1
	n.prepareForScroll()

	delay := 10 * time.Millisecond
	if JC.IsMobile {
		delay = 50 * time.Millisecond
	}

	JC.MainDebouncer.Cancel("layout_update_" + n.uuid)
	JC.MainDebouncer.Call("layout_update_"+n.uuid, delay, func() {
		fyne.Do(func() {
			scaledHeight := n.scaledItemHeight * float32(n.itemTotal)
			if n.scaledHeight != scaledHeight {
				n.scaledHeight = scaledHeight
				n.scrollContent.SetMinSize(fyne.NewSize(1, n.scaledHeight))
			}

			if n.scrollBox.Offset.Y != 0 {
				n.scrollBox.ScrollToTop()
			}

			n.refreshContent()
		})
	})
}

func (n *completionList) Resize(size fyne.Size) {
	n.lastSize = size
	n.BaseWidget.Resize(size)
}

func (n *completionList) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (n *completionList) OnSelected(index int) {
	if index >= 0 && index < n.itemTotal {
		n.onChange(n.data[index])
	}
}

func (n *completionList) TypedKey(event *fyne.KeyEvent) {

	if n.itemTotal == 0 {
		n.selected = -1
		n.scrollOffset = 0
		n.refreshContent()
		return
	}

	switch event.Name {
	case fyne.KeyDown:
		if n.selected < n.itemTotal-1 && n.selected >= 0 {
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
			n.selected = n.itemTotal - 1
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

	if n.scrollOffset < 0 {
		n.scrollOffset = 0
	}
	if n.scrollOffset > n.maxOffset {
		n.scrollOffset = n.maxOffset
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
		n.done = make(chan struct{})

		go func() {
			ticker := time.NewTicker(n.fps)
			defer ticker.Stop()

			lastY := ev.Position.Y

			for {
				select {
				case <-n.done:
					return
				case <-ticker.C:
					currentY := n.position.Y
					deltaY := currentY - lastY

					if math.Abs(float64(deltaY)) < n.scrollLimiter {
						continue
					}

					scrollEvent := &fyne.ScrollEvent{
						Scrolled: fyne.Delta{
							DX: 0,
							DY: -deltaY,
						},
					}

					fyne.Do(func() {
						n.Scrolled(scrollEvent)
					})

					lastY = currentY
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

func (n *completionList) prepareForScroll() {

	limiter := float32(0.3)

	n.itemTotal = len(n.data)
	n.scaledItemHeight = n.itemHeight

	oh := n.itemHeight * float32(n.itemTotal)
	mx := float32(1000)

	if JC.IsMobile {
		mx = 500
	}

	if oh > mx {
		n.scaledItemHeight = float32(math.Ceil(float64(mx/float32(n.itemTotal) + float32(n.itemVisible))))
	}

	if n.itemTotal > 500 {
		limiter = 0.1
	}

	n.scrollLimiter = float64(n.scaledItemHeight * limiter)
	n.maxOffset = n.itemTotal - n.itemVisible
}

func (n *completionList) refreshContent() {
	for i, obj := range n.contentBox.Objects {
		dataIndex := n.scrollOffset + i

		item, ok := obj.(*completionText)
		if !ok {
			continue
		}
		if dataIndex >= 0 && dataIndex < n.itemTotal {
			item.SetText(n.data[dataIndex])
			item.SetIndex(dataIndex)
		}
	}
}

func (n *completionList) scrollingContent(offset fyne.Position) {
	newOffset := int(offset.Y / n.scaledItemHeight)

	if newOffset > n.maxOffset {
		newOffset = n.maxOffset
	}
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset != n.scrollOffset {
		n.scrollOffset = newOffset
		n.refreshContent()
	}
}

func (n *completionList) Tapped(_ *fyne.PointEvent) {
}

func (n *completionList) TypedRune(r rune) {
}

func (n *completionList) FocusGained() {
}

func (n *completionList) FocusLost() {
}
