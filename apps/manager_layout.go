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
	state             int
	lastDisplayUpdate time.Time
}

func (m *layoutManager) Init() {
	m.mu = sync.RWMutex{}
	m.state = -9
	m.lastDisplayUpdate = time.Time{}
}

func (m *layoutManager) RefreshLayout() {

	m.mu.RLock()
	scroll := m.scroll
	m.mu.RUnlock()

	if scroll != nil {
		scroll.Refresh()
	}
}

func (m *layoutManager) UpdateState() {
	if m == nil {
		return
	}

	m.mu.Lock()
	currentState := m.state
	content := m.content
	scroll := m.scroll
	m.mu.Unlock()

	if content == nil || !UseStatus().IsReady() {
		m.mu.Lock()
		scroll.Content = m.loading
		m.state = -1
		m.mu.Unlock()
	} else if !UseStatus().IsValidConfig() && !UseStatus().IsValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !UseStatus().IsValidCrypto() && !UseStatus().IsValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !UseStatus().IsValidPanels() {
		m.mu.Lock()
		scroll.Content = m.actionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !UseStatus().HasError() || UseStatus().IsValidPanels() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = UseStatus().PanelsCount()
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
		if !UseStatus().IsValidCrypto() || !UseStatus().IsTickerShown() {
			m.setTickers(container.NewWithoutLayout())
			return
		}

		m.mu.RLock()
		tickers := m.tickers
		populated := m.tickersPopulated
		m.mu.RUnlock()

		if tickers != nil && tickers != populated && UseStatus().IsValidTickers() {
			m.setTickers(m.tickersPopulated)
			return
		}

		if tickers != nil && tickers == populated && !UseStatus().IsValidTickers() {
			m.setTickers(container.NewWithoutLayout())
			return
		}

		if tickers == nil && UseStatus().IsValidTickers() {
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

func (m *layoutManager) RemoveOverlay(container *fyne.Container) {
	m.container.Remove(container)
	m.mu.Lock()
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = nil
	}
	m.mu.Unlock()
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
	m.container.Add(container)
	m.mu.Lock()
	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.overlay = container
	}
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

func (m *layoutManager) setTickers(container *fyne.Container) {
	if container == nil {
		return
	}

	m.mu.Lock()
	m.tickers = container
	m.container.Objects[2] = container

	if layout, ok := m.container.Layout.(*mainLayout); ok {
		layout.tickers = container
	}
	m.mu.Unlock()

	m.tickers.Refresh()

	m.container.Refresh()
}

func NewAppLayout() fyne.CanvasObject {

	JW.NotificationInit()

	manager := UseLayout()
	manager.topBar = NewTopBar()

	manager.loading = NewAppPage(nil, "Loading...", nil)
	manager.error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.actionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		UseAction().Call(JC.ACT_PANEL_ADD)
	})

	settingIcon := theme.SettingsIcon()
	manager.actionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		UseAction().Call(JC.ACT_OPEN_SETTINGS)
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.actionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		UseAction().Call(JC.ACT_CRYPTO_REFRESH_MAP)
	})

	manager.scroll = container.NewVScroll(nil)
	manager.UpdateState()

	manager.tickers = container.NewWithoutLayout()

	manager.placeholder = NewDragPlaceholder()

	layout := &mainLayout{
		padding:     JC.UseTheme().Size(JC.SizeLayoutPadding),
		parent:      manager,
		topBar:      manager.topBar,
		tickers:     manager.tickers,
		content:     manager.scroll,
		placeholder: manager.placeholder,
		background:  canvas.NewRectangle(JC.UseTheme().GetColor(theme.ColorNameBackground)),
		overlay:     nil,
	}

	manager.container = container.New(
		layout,
		layout.background,
		layout.topBar,
		layout.tickers,
		layout.content,
		layout.placeholder,
	)

	return fynetooltip.AddWindowToolTipLayer(
		manager.container,
		JC.Window.Canvas())
}

func RegisterLayoutManager() *layoutManager {
	if layoutManagerStorage == nil {
		layoutManagerStorage = &layoutManager{}
	}
	return layoutManagerStorage
}

func UseLayout() *layoutManager {
	return layoutManagerStorage
}
