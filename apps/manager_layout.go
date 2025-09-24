package apps

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
	JW "jxwatcher/widgets"
)

var layoutManagerStorage *layoutManager = nil

type layoutManager struct {
	mu                sync.RWMutex
	topBar            *fyne.Container
	content           *fyne.CanvasObject
	tickers           *fyne.Container
	tickersPopulated  *fyne.Container
	scroll            *container.Scroll
	container         *fyne.Container
	placeholder       *dragPlaceholder
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

func (m *layoutManager) Init() {
	m.mu = sync.RWMutex{}
	// m.topBar = container.NewWithoutLayout()
	// m.content = nil
	// m.tickers = container.NewWithoutLayout()
	// m.tickersPopulated = container.NewWithoutLayout()
	// m.scroll = container.NewScroll(nil)
	// m.container = container.NewWithoutLayout()
	// m.dragPlaceholder = nil

	// m.actionAddPanel = nil
	// m.actionFixSetting = nil
	// m.actionGetCryptos = nil
	// m.loading = nil
	// m.error = nil

	m.maxOffset = 0
	m.contentTopY = 0
	m.contentBottomY = 0
	m.state = 9
	m.lastDisplayUpdate = time.Time{}
	m.contentWidth = 0
	m.contentHeight = 0
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

	if content == nil || !UseStatus().IsReady() {
		m.mu.Lock()
		scroll.Content = m.loading
		m.state = -1
		m.mu.Unlock()
	} else if !UseStatus().ValidConfig() {
		m.mu.Lock()
		scroll.Content = m.actionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !UseStatus().ValidCryptos() {
		m.mu.Lock()
		scroll.Content = m.actionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !UseStatus().ValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !UseStatus().HasError() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = UseStatus().panels_count
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

	if UseStatus().IsReady() {
		if !UseStatus().ValidCryptos() {
			m.setTickers(container.NewWithoutLayout())
			return
		}

		m.mu.RLock()
		tickers := m.tickers
		populated := m.tickersPopulated
		m.mu.RUnlock()

		if tickers != nil && tickers != populated && UseStatus().ValidTickers() {
			m.setTickers(m.tickersPopulated)
			return
		}

		if tickers != nil && tickers == populated && !UseStatus().ValidTickers() {
			m.setTickers(container.NewWithoutLayout())
			return
		}

		if tickers == nil && UseStatus().ValidTickers() {
			m.setTickers(populated)
			return
		}
	}
}

func (m *layoutManager) GetDisplayUpdate() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastDisplayUpdate
}

func (m *layoutManager) GetContentHeight() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentHeight
}

func (m *layoutManager) GetContentTopY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentTopY
}

func (m *layoutManager) GetContentBottomY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentBottomY
}

func (m *layoutManager) RemoveOverlay(container *fyne.Container) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.container.Remove(container)
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = nil
	}
}

func (m *layoutManager) RegisterDisplayUpdate(ts time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastDisplayUpdate = ts
}

func (m *layoutManager) RegisterContent(container fyne.CanvasObject) {
	m.mu.Lock()
	m.content = &container
	m.mu.Unlock()
}

func (m *layoutManager) RegisterOverlay(container *fyne.Container) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.container.Add(container)
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = container
	}
}

func (m *layoutManager) RegisterTickers(container *fyne.Container) {
	if container == nil {
		return
	}

	m.mu.Lock()
	m.tickersPopulated = container
	m.mu.Unlock()
}

func (m *layoutManager) UseContainer() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container
}

func (m *layoutManager) UseScroll() *container.Scroll {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll
}

func (m *layoutManager) UsePlaceholder() *dragPlaceholder {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.placeholder
}

func (m *layoutManager) ScrollBy(delta float32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	scroll := m.scroll
	if scroll == nil || scroll.Content == nil {
		return
	}

	current := scroll.Offset.Y
	newOffset := current + delta

	if m.maxOffset == -1 {
		contentHeight := scroll.Content.MinSize().Height
		viewportHeight := scroll.Size().Height

		if contentHeight <= viewportHeight {
			m.maxOffset = 0
		} else {
			m.maxOffset = contentHeight - viewportHeight
		}
	}

	max := m.maxOffset

	if newOffset < 0 {
		if current <= 0 {
			return
		}
		newOffset = 0
	} else if newOffset > max {
		if current >= max {
			return
		}
		newOffset = max
	}

	if scroll.Offset.Y == newOffset {
		return
	}

	scroll.Offset.Y = newOffset
	m.RefreshLayout()
}

func (m *layoutManager) setTickers(container *fyne.Container) {
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

	JC.UseDebouncer().Call("refreshing_layout_container", 5*time.Millisecond, func() {
		m.mu.RLock()
		fyne.Do(m.container.Refresh)
		m.mu.RUnlock()
	})
}

func (m *layoutManager) setContentSize(width, height float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contentWidth = width
	m.contentHeight = height
}

func (m *layoutManager) setMaxOffset(val float32) {
	m.mu.Lock()
	m.maxOffset = val
	m.mu.Unlock()
}

func (m *layoutManager) setContentTopY(val float32) {
	m.mu.Lock()
	m.contentTopY = val
	m.mu.Unlock()
}

func (m *layoutManager) setContentBottomY(val float32) {
	m.mu.Lock()
	m.contentBottomY = val
	m.mu.Unlock()
}

func NewAppLayout() fyne.CanvasObject {

	JW.NotificationInit()

	manager := UseLayout()
	manager.topBar = NewTopBar()

	manager.loading = NewAppPage(nil, "Loading...", nil)
	manager.error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.actionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		UseAction().Call("add_panel")
	})

	settingIcon := theme.SettingsIcon()
	manager.actionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		UseAction().Call("open_settings")
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.actionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		UseAction().Call("refresh_cryptos")
	})

	manager.scroll = container.NewVScroll(nil)
	manager.Refresh()

	manager.tickers = container.NewWithoutLayout()

	manager.placeholder = NewDragPlaceholder()

	layout := &mainLayout{
		padding:     10,
		parent:      manager,
		topBar:      manager.topBar,
		tickers:     manager.tickers,
		content:     manager.scroll,
		placeholder: manager.placeholder,
		overlay:     nil,
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

func RegisteLayoutManager() *layoutManager {
	if layoutManagerStorage == nil {
		layoutManagerStorage = &layoutManager{}
	}
	return layoutManagerStorage
}

func UseLayout() *layoutManager {
	return layoutManagerStorage
}
