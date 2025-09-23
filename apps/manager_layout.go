package apps

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
	JW "jxwatcher/widgets"
)

var LayoutManager *layoutManager = nil
var DragPlaceholder fyne.CanvasObject

type layoutManager struct {
	mu                sync.RWMutex
	topBar            *fyne.Container
	content           *fyne.CanvasObject
	tickers           *fyne.Container
	tickersPopulated  *fyne.Container
	scroll            *container.Scroll
	container         *fyne.Container
	actionAddPanel    *staticPage
	actionFixSetting  *staticPage
	actionGetCryptos  *staticPage
	loading           *staticPage
	error             *staticPage
	maxOffset         float32
	contentTopY       float32
	contentBottomY    float32
	state             int
	lastDisplayUpdate time.Time
	contentWidth      float32
	contentHeight     float32
}

func (m *layoutManager) TopBar() fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topBar
}

func (m *layoutManager) Content() *fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content
}

func (m *layoutManager) Tickers() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tickers
}

func (m *layoutManager) Scroll() *container.Scroll {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll
}

func (m *layoutManager) Container() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container
}

func (m *layoutManager) ContainerSize() fyne.Size {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container.Size()
}

func (m *layoutManager) ActionAddPanel() *staticPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionAddPanel
}

func (m *layoutManager) ActionFixSetting() *staticPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionFixSetting
}

func (m *layoutManager) ActionGetCryptos() *staticPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionGetCryptos
}

func (m *layoutManager) LoadingPage() *staticPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loading
}

func (m *layoutManager) ErrorPage() *staticPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.error
}

func (m *layoutManager) MaxOffset() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxOffset
}

func (m *layoutManager) ContentTopY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentTopY
}

func (m *layoutManager) ContentBottomY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentBottomY
}

func (m *layoutManager) State() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *layoutManager) OffsetY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.Y
}

func (m *layoutManager) OffsetX() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.X
}

func (m *layoutManager) Height() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Height
}

func (m *layoutManager) Width() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Width
}

func (m *layoutManager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll != nil && m.content != nil
}

func (m *layoutManager) SetMaxOffset(val float32) {
	m.mu.Lock()
	m.maxOffset = val
	m.mu.Unlock()
}

func (m *layoutManager) SetContentTopY(val float32) {
	m.mu.Lock()
	m.contentTopY = val
	m.mu.Unlock()
}

func (m *layoutManager) SetContentBottomY(val float32) {
	m.mu.Lock()
	m.contentBottomY = val
	m.mu.Unlock()
}

func (m *layoutManager) SetPage(container fyne.CanvasObject) {
	m.mu.Lock()
	m.content = &container
	m.mu.Unlock()
}

func (m *layoutManager) RegisterTickers(container *fyne.Container) {
	if container == nil {
		return
	}

	m.mu.Lock()
	m.tickersPopulated = container
	m.mu.Unlock()
}

func (m *layoutManager) SetTickers(container *fyne.Container) {
	if container == nil {
		return
	}

	m.mu.Lock()
	m.tickers = container
	m.container.Objects[1] = container

	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.tickers = container
	}
	m.mu.Unlock()

	m.tickers.Refresh()
	m.RefreshContainer()
}

func (m *layoutManager) SetOffsetY(offset float32) {
	m.mu.Lock()
	if m.scroll == nil || m.scroll.Offset.Y == offset {
		m.mu.Unlock()
		return
	}
	m.scroll.Offset.Y = offset
	m.mu.Unlock()
	m.RefreshLayout()
}

func (m *layoutManager) ScrollBy(delta float32) {
	current := m.OffsetY()
	newOffset := current + delta

	m.mu.RLock()
	max := m.maxOffset
	m.mu.RUnlock()

	if max == -1 {
		m.ComputeMaxScrollOffset()
		m.mu.RLock()
		max = m.maxOffset
		m.mu.RUnlock()
	}

	if newOffset < 0 {
		if current > 0 {
			newOffset = 0
		} else {
			return
		}
	} else if newOffset > max {
		if current < max {
			newOffset = max
		} else {
			return
		}
	}

	m.SetOffsetY(newOffset)
}

func (m *layoutManager) ComputeMaxScrollOffset() {
	m.mu.RLock()
	scroll := m.scroll
	m.mu.RUnlock()

	if scroll == nil || scroll.Content == nil {
		return
	}

	contentHeight := scroll.Content.MinSize().Height
	viewportHeight := scroll.Size().Height

	m.mu.Lock()
	if contentHeight <= viewportHeight {
		m.maxOffset = 0
	} else {
		m.maxOffset = contentHeight - viewportHeight
	}
	m.mu.Unlock()
}

func (m *layoutManager) RefreshLayout() {
	JC.UseDebouncer().Call("refreshing_layout_layout", 5*time.Millisecond, func() {
		m.mu.RLock()
		if m.scroll != nil {
			fyne.Do(m.scroll.Refresh)
		}
		m.mu.RUnlock()
	})
}

func (m *layoutManager) RefreshContainer() {
	JC.UseDebouncer().Call("refreshing_layout_container", 5*time.Millisecond, func() {
		m.mu.RLock()
		fyne.Do(m.container.Refresh)
		m.mu.RUnlock()
	})
}

func (m *layoutManager) Refresh() {
	if m == nil {
		return
	}

	m.mu.Lock()
	m.maxOffset = -1
	currentState := m.state
	content := m.content
	scroll := m.scroll
	m.mu.Unlock()

	if content == nil || !StatusManager.IsReady() {
		m.mu.Lock()
		scroll.Content = m.loading
		m.state = -1
		m.mu.Unlock()
	} else if !StatusManager.ValidConfig() {
		m.mu.Lock()
		scroll.Content = m.actionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !StatusManager.ValidCryptos() {
		m.mu.Lock()
		scroll.Content = m.actionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !StatusManager.ValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !StatusManager.HasError() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = StatusManager.panels_count
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		scroll.Content = m.error
		m.state = -5
		m.mu.Unlock()
	}

	m.mu.RLock()
	stateChanged := m.state != currentState
	m.mu.RUnlock()

	if stateChanged {
		m.RefreshLayout()
	}

	if StatusManager.IsReady() {
		if !StatusManager.ValidCryptos() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		m.mu.RLock()
		tickers := m.tickers
		populated := m.tickersPopulated
		m.mu.RUnlock()

		if tickers != nil && tickers != populated && StatusManager.ValidTickers() {
			m.SetTickers(m.tickersPopulated)
			return
		}

		if tickers != nil && tickers == populated && !StatusManager.ValidTickers() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		if tickers == nil && StatusManager.ValidTickers() {
			m.SetTickers(populated)
			return
		}
	}
}

func (m *layoutManager) AddToContainer(container *fyne.Container) {
	m.container.Add(container)
}

func (m *layoutManager) RemoveFromContainer(container *fyne.Container) {
	m.container.Remove(container)
}

func (m *layoutManager) SetOverlay(container *fyne.Container) {
	m.container.Add(container)
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = container
	}
}

func (m *layoutManager) RemoveOverlay(container *fyne.Container) {
	m.container.Remove(container)
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = nil
	}
}

func (lm *layoutManager) SetLastDisplayUpdate(ts time.Time) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.lastDisplayUpdate = ts
}

func (lm *layoutManager) GetLastDisplayUpdate() time.Time {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.lastDisplayUpdate
}

func (lm *layoutManager) SetContentSize(width, height float32) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.contentWidth = width
	lm.contentHeight = height
}

func (lm *layoutManager) GetContentWidth() float32 {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.contentWidth
}

func (lm *layoutManager) GetContentHeight() float32 {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.contentHeight
}

func NewAppLayout() fyne.CanvasObject {

	JW.NotificationInit()

	manager := &layoutManager{
		topBar: NewTopBar(),
	}

	LayoutManager = manager

	manager.loading = NewAppPage(nil, "Loading...", nil)
	manager.error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.actionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		UseActionManager().Call("add_panel")
	})

	settingIcon := theme.SettingsIcon()
	manager.actionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		UseActionManager().Call("open_settings")
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.actionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		UseActionManager().Call("refresh_cryptos")
	})

	manager.scroll = container.NewVScroll(nil)
	manager.Refresh()

	manager.tickers = container.NewWithoutLayout()

	layout := &mainLayout{
		padding:     10,
		parent:      manager,
		topBar:      manager.topBar,
		tickers:     manager.tickers,
		content:     manager.scroll,
		placeholder: nil,
		overlay:     nil,
	}

	DragPlaceholder = canvas.NewRectangle(JC.ThemeColor(JC.ColorNameTransparent))
	if rect, ok := DragPlaceholder.(*canvas.Rectangle); ok {
		rect.CornerRadius = JC.ThemeSize(JC.SizePanelBorderRadius)
		layout.placeholder = rect
	}

	manager.container = container.New(
		layout,
		layout.topBar,
		layout.tickers,
		layout.content,
		layout.placeholder,
	)

	return fynetooltip.AddWindowToolTipLayer(
		manager.container,
		JC.Window.Canvas())
}
