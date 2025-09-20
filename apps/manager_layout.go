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

var AppLayout *appLayout = nil
var DragPlaceholder fyne.CanvasObject

type appLayout struct {
	mu               sync.RWMutex
	topBar           *fyne.Container
	content          *fyne.CanvasObject
	tickers          *fyne.Container
	scroll           *container.Scroll
	container        *fyne.Container
	actionAddPanel   *appPage
	actionFixSetting *appPage
	actionGetCryptos *appPage
	loading          *appPage
	error            *appPage
	maxOffset        float32
	contentTopY      float32
	contentBottomY   float32
	state            int
}

func (m *appLayout) TopBar() fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topBar
}

func (m *appLayout) Content() *fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content
}

func (m *appLayout) Tickers() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tickers
}

func (m *appLayout) Scroll() *container.Scroll {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll
}

func (m *appLayout) Container() *fyne.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container
}

func (m *appLayout) ContainerSize() fyne.Size {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.container.Size()
}

func (m *appLayout) ActionAddPanel() *appPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionAddPanel
}

func (m *appLayout) ActionFixSetting() *appPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionFixSetting
}

func (m *appLayout) ActionGetCryptos() *appPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionGetCryptos
}

func (m *appLayout) LoadingPage() *appPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loading
}

func (m *appLayout) ErrorPage() *appPage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.error
}

func (m *appLayout) MaxOffset() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxOffset
}

func (m *appLayout) ContentTopY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentTopY
}

func (m *appLayout) ContentBottomY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contentBottomY
}

func (m *appLayout) State() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *appLayout) OffsetY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.Y
}

func (m *appLayout) OffsetX() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return 0
	}
	return m.scroll.Offset.X
}

func (m *appLayout) Height() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Height
}

func (m *appLayout) Width() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.scroll == nil {
		return -1
	}
	return m.scroll.Size().Width
}

func (m *appLayout) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scroll != nil && m.content != nil
}

func (m *appLayout) SetMaxOffset(val float32) {
	m.mu.Lock()
	m.maxOffset = val
	m.mu.Unlock()
}

func (m *appLayout) SetContentTopY(val float32) {
	m.mu.Lock()
	m.contentTopY = val
	m.mu.Unlock()
}

func (m *appLayout) SetContentBottomY(val float32) {
	m.mu.Lock()
	m.contentBottomY = val
	m.mu.Unlock()
}

func (m *appLayout) SetPage(container fyne.CanvasObject) {
	m.mu.Lock()
	m.content = &container
	m.mu.Unlock()
}

func (m *appLayout) SetTickers(container *fyne.Container) {
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

func (m *appLayout) SetOffsetY(offset float32) {
	m.mu.Lock()
	if m.scroll == nil || m.scroll.Offset.Y == offset {
		m.mu.Unlock()
		return
	}
	m.scroll.Offset.Y = offset
	m.mu.Unlock()
	m.RefreshLayout()
}

func (m *appLayout) ScrollBy(delta float32) {
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

func (m *appLayout) ComputeMaxScrollOffset() {
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

func (m *appLayout) RefreshLayout() {
	JC.MainDebouncer.Call("refreshing_layout_layout", 5*time.Millisecond, func() {
		m.mu.RLock()
		if m.scroll != nil {
			fyne.Do(m.scroll.Refresh)
		}
		m.mu.RUnlock()
	})
}

func (m *appLayout) RefreshContainer() {
	JC.MainDebouncer.Call("refreshing_layout_container", 5*time.Millisecond, func() {
		m.mu.RLock()
		fyne.Do(m.container.Refresh)
		m.mu.RUnlock()
	})
}

func (m *appLayout) Refresh() {
	if m == nil {
		return
	}

	m.mu.Lock()
	m.maxOffset = -1
	currentState := m.state
	content := m.content
	scroll := m.scroll
	m.mu.Unlock()

	if content == nil || !AppStatus.IsReady() {
		m.mu.Lock()
		scroll.Content = m.loading
		m.state = -1
		m.mu.Unlock()
	} else if !AppStatus.ValidConfig() {
		m.mu.Lock()
		scroll.Content = m.actionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !AppStatus.ValidCryptos() {
		m.mu.Lock()
		scroll.Content = m.actionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !AppStatus.ValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !AppStatus.HasError() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = AppStatus.panels_count
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

	if AppStatus.IsReady() {
		if !AppStatus.ValidCryptos() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		m.mu.RLock()
		tickers := m.tickers
		m.mu.RUnlock()

		if tickers != nil && tickers != JC.Tickers && AppStatus.ValidTickers() {
			m.SetTickers(JC.Tickers)
			return
		}

		if tickers != nil && tickers == JC.Tickers && !AppStatus.ValidTickers() {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		if tickers == nil && AppStatus.ValidTickers() {
			m.SetTickers(JC.Tickers)
			return
		}
	}
}

func (m *appLayout) AddToContainer(container *fyne.Container) {
	m.container.Add(container)
}

func (m *appLayout) RemoveFromContainer(container *fyne.Container) {
	m.container.Remove(container)
}

func (m *appLayout) SetOverlay(container *fyne.Container) {
	m.container.Add(container)
	if layout, ok := m.container.Layout.(*AppMainLayout); ok {
		layout.overlay = container
	}
}

func (m *appLayout) RemoveOverlay(container *fyne.Container) {
	m.container.Remove(container)
	if layout, ok := m.container.Layout.(*AppMainLayout); ok {
		layout.overlay = nil
	}
}

func NewAppLayout() fyne.CanvasObject {
	JC.NotificationContainer = JW.NewNotificationDisplay()

	manager := &appLayout{
		topBar: NewTopBar(),
	}

	AppLayout = manager

	manager.loading = NewAppPage(nil, "Loading...", nil)
	manager.error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.actionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		AppActions.CallButton("add_panel")
	})

	settingIcon := theme.SettingsIcon()
	manager.actionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		AppActions.CallButton("open_settings")
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.actionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		AppActions.CallButton("refresh_cryptos")
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
