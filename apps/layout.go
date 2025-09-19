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

var AppLayoutManager *AppLayout = nil
var DragPlaceholder fyne.CanvasObject

type AppLayout struct {
	mu               sync.RWMutex
	topBar           *fyne.Container
	content          *fyne.CanvasObject
	tickers          *fyne.Container
	scroll           *container.Scroll
	container        *fyne.Container
	actionAddPanel   *AppPage
	actionFixSetting *AppPage
	actionGetCryptos *AppPage
	loading          *AppPage
	error            *AppPage
	maxOffset        float32
	contentTopY      float32
	contentBottomY   float32
	state            int
}

func (m *AppLayout) TopBar() fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topBar
}

func (m *AppLayout) Content() *fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content
}

func (m *AppLayout) Tickers() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tickers
}

func (m *AppLayout) Scroll() *container.Scroll {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll
}

func (m *AppLayout) Container() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container
}

func (m *AppLayout) ContainerSize() fyne.Size {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container.Size()
}

func (m *AppLayout) ActionAddPanel() *AppPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionAddPanel
}

func (m *AppLayout) ActionFixSetting() *AppPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionFixSetting
}

func (m *AppLayout) ActionGetCryptos() *AppPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionGetCryptos
}

func (m *AppLayout) LoadingPage() *AppPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loading
}

func (m *AppLayout) ErrorPage() *AppPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.error
}

func (m *AppLayout) MaxOffset() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxOffset
}

func (m *AppLayout) ContentTopY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentTopY
}

func (m *AppLayout) ContentBottomY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentBottomY
}

func (m *AppLayout) State() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *AppLayout) OffsetY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.Y
}

func (m *AppLayout) OffsetX() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.X
}

func (m *AppLayout) Height() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Height
}

func (m *AppLayout) Width() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Width
}

func (m *AppLayout) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll != nil && m.content != nil
}

func (m *AppLayout) SetMaxOffset(val float32) {
	m.mu.Lock()
	m.maxOffset = val
	m.mu.Unlock()
}

func (m *AppLayout) SetContentTopY(val float32) {
	m.mu.Lock()
	m.contentTopY = val
	m.mu.Unlock()
}

func (m *AppLayout) SetContentBottomY(val float32) {
	m.mu.Lock()
	m.contentBottomY = val
	m.mu.Unlock()
}

func (m *AppLayout) SetPage(container fyne.CanvasObject) {
	m.mu.Lock()
	m.content = &container
	m.mu.Unlock()
}

func (m *AppLayout) SetTickers(container *fyne.Container) {
	if container == nil {
		return
	}

	m.mu.Lock()
	m.tickers = container
	m.container.Objects[1] = container

	if layout, ok := m.container.Layout.(*AppMainLayout); ok {
		layout.tickers = container
	}
	m.mu.Unlock()

	m.tickers.Refresh()
	m.RefreshContainer()
}

func (m *AppLayout) SetOffsetY(offset float32) {
	m.mu.Lock()
	if m.scroll == nil || m.scroll.Offset.Y == offset {
		m.mu.Unlock()
		return
	}
	m.scroll.Offset.Y = offset
	m.mu.Unlock()
	m.RefreshLayout()
}

func (m *AppLayout) ScrollBy(delta float32) {
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

func (m *AppLayout) ComputeMaxScrollOffset() {
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

func (m *AppLayout) RefreshLayout() {
	JC.MainDebouncer.Call("refreshing_layout_layout", 5*time.Millisecond, func() {
		m.mu.RLock()
		if m.scroll != nil {
			fyne.Do(m.scroll.Refresh)
		}
		m.mu.RUnlock()
	})
}

func (m *AppLayout) RefreshContainer() {
	JC.MainDebouncer.Call("refreshing_layout_container", 5*time.Millisecond, func() {
		m.mu.RLock()
		fyne.Do(m.container.Refresh)
		m.mu.RUnlock()
	})
}

func (m *AppLayout) Refresh() {
	if m == nil {
		return
	}

	m.mu.Lock()
	m.maxOffset = -1
	currentState := m.state
	content := m.content
	scroll := m.scroll
	m.mu.Unlock()

	if content == nil || !AppStatusManager.IsReady() {
		m.mu.Lock()
		scroll.Content = m.loading
		m.state = -1
		m.mu.Unlock()
	} else if !AppStatusManager.ValidConfig() {
		m.mu.Lock()
		scroll.Content = m.actionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !AppStatusManager.ValidCryptos() {
		m.mu.Lock()
		scroll.Content = m.actionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !AppStatusManager.ValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !AppStatusManager.HasError() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = AppStatusManager.panels_count
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

	if AppStatusManager.IsReady() {
		if !AppStatusManager.ValidCryptos() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		m.mu.RLock()
		tickers := m.tickers
		m.mu.RUnlock()

		if tickers != nil && tickers != JC.Tickers && AppStatusManager.ValidTickers() {
			m.SetTickers(JC.Tickers)
			return
		}

		if tickers != nil && tickers == JC.Tickers && !AppStatusManager.ValidTickers() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		if tickers == nil && AppStatusManager.ValidTickers() {
			m.SetTickers(JC.Tickers)
			return
		}
	}
}

func (m *AppLayout) AddToContainer(container *fyne.Container) {
	m.container.Add(container)
}

func (m *AppLayout) RemoveFromContainer(container *fyne.Container) {
	m.container.Remove(container)
}

func (m *AppLayout) SetOverlay(container *fyne.Container) {
	m.container.Add(container)
	if layout, ok := m.container.Layout.(*AppMainLayout); ok {
		layout.overlay = container
	}
}

func (m *AppLayout) RemoveOverlay(container *fyne.Container) {
	m.container.Remove(container)
	if layout, ok := m.container.Layout.(*AppMainLayout); ok {
		layout.overlay = nil
	}
}

func NewAppLayoutManager() fyne.CanvasObject {
	JC.NotificationContainer = JW.NewNotificationDisplay()

	manager := &AppLayout{
		topBar: NewTopBar(),
	}

	AppLayoutManager = manager

	manager.loading = NewAppPage(nil, "Loading...", nil)
	manager.error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.actionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		AppActionManager.CallButton("add_panel")
	})

	settingIcon := theme.SettingsIcon()
	manager.actionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		AppActionManager.CallButton("open_settings")
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.actionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		AppActionManager.CallButton("refresh_cryptos")
	})

	manager.scroll = container.NewVScroll(nil)
	manager.Refresh()

	manager.tickers = container.NewWithoutLayout()

	layout := &AppMainLayout{
		padding:     10,
		topBar:      manager.topBar,
		tickers:     manager.tickers,
		content:     manager.scroll,
		placeholder: nil,
		overlay:     nil, // set this if you have an overlay object
	}

	DragPlaceholder = canvas.NewRectangle(JC.Transparent)
	if rect, ok := DragPlaceholder.(*canvas.Rectangle); ok {
		rect.CornerRadius = JC.PanelBorderRadius
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
