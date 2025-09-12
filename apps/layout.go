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

type AppMainLayout struct {
	Padding float32
}

var AppLayoutManager *AppLayout = nil
var DragPlaceholder fyne.CanvasObject

func (a *AppMainLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	// Background setup
	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(size)
	bg.Resize(size)

	topBar := objects[0]
	topHeight := topBar.MinSize().Height
	contentY := topHeight + 2*a.Padding

	newContentWidth := size.Width - 2*a.Padding
	newContentHeight := size.Height - contentY - 2*a.Padding

	// TopBar layout
	newTopBarPos := fyne.NewPos(a.Padding, a.Padding)
	newTopBarSize := fyne.NewSize(newContentWidth, topHeight)

	if topBar.Position() != newTopBarPos || topBar.Size() != newTopBarSize {
		topBar.Move(newTopBarPos)
		topBar.Resize(newTopBarSize)
	}

	// Tickers Layout
	tickers, ok := objects[1].(*fyne.Container)
	if ok && len(tickers.Objects) > 0 {
		tickerHeight := tickers.MinSize().Height
		newTickersPos := fyne.NewPos(a.Padding, contentY)
		newTickersSize := fyne.NewSize(newContentWidth, tickerHeight)

		if tickers.Position() != newTickersPos || tickers.Size() != newTickersSize {
			tickers.Move(newTickersPos)
			tickers.Resize(newTickersSize)
		}

		contentY += tickerHeight
		newContentHeight -= tickerHeight
	}

	// Content layout
	content := objects[2]
	newContentPos := fyne.NewPos(a.Padding, contentY)
	newContentSize := fyne.NewSize(newContentWidth, newContentHeight)

	if content.Position() != newContentPos || content.Size() != newContentSize {
		content.Move(newContentPos)
		content.Resize(newContentSize)
	}

	placeholder := objects[3]
	placeholder.Move(fyne.NewPos(0, -JC.PanelHeight))

	// Update global layout dimensions
	JC.MainLayoutContentWidth = newContentWidth
	JC.MainLayoutContentHeight = newContentHeight

	AppLayoutManager.MaxOffset = -1
	AppLayoutManager.ContentTopY = newContentPos.Y
	AppLayoutManager.ContentBottomY = newContentPos.Y + newContentHeight

}

func (a *AppMainLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	top := objects[0].MinSize()
	content := objects[2].MinSize()

	width := fyne.Max(top.Width, content.Width) + 2*a.Padding
	height := top.Height + content.Height + 3*a.Padding

	tickers, ok := objects[1].(*fyne.Container)
	if ok && len(tickers.Objects) > 0 {
		height += tickers.MinSize().Height
	}

	return fyne.NewSize(width, height)
}

type AppLayout struct {
	mu               sync.RWMutex
	TopBar           fyne.CanvasObject
	Content          *fyne.CanvasObject
	Tickers          *fyne.Container
	Scroll           *container.Scroll
	Container        fyne.Container
	ActionAddPanel   *AppPage
	ActionFixSetting *AppPage
	ActionGetCryptos *AppPage
	Loading          *AppPage
	Error            *AppPage
	MaxOffset        float32
	ContentTopY      float32
	ContentBottomY   float32
	state            int
}

func (m *AppLayout) OffsetY() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Scroll == nil {
		return 0
	}
	return m.Scroll.Offset.Y
}

func (m *AppLayout) OffsetX() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Scroll == nil {
		return 0
	}
	return m.Scroll.Offset.X
}

func (m *AppLayout) Height() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Scroll == nil {
		return -1
	}
	return m.Scroll.Size().Height
}

func (m *AppLayout) Width() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Scroll == nil {
		return -1
	}
	return m.Scroll.Size().Width
}

func (m *AppLayout) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Scroll != nil && m.Content != nil
}

func (m *AppLayout) SetPage(container fyne.CanvasObject) {
	m.mu.Lock()
	m.Content = &container
	m.mu.Unlock()
}

func (m *AppLayout) SetTickers(container *fyne.Container) {
	if container == nil {
		return
	}
	m.mu.Lock()
	m.Tickers = container
	m.Container.Objects[1] = container
	m.mu.Unlock()
	m.RefreshContainer()
}

func (m *AppLayout) SetOffsetY(offset float32) {
	m.mu.Lock()
	if m.Scroll == nil || m.Scroll.Offset.Y == offset {
		m.mu.Unlock()
		return
	}
	m.Scroll.Offset.Y = offset
	m.mu.Unlock()
	m.RefreshLayout()
}

func (m *AppLayout) ScrollBy(delta float32) {
	current := m.OffsetY()
	newOffset := current + delta

	m.mu.RLock()
	max := m.MaxOffset
	m.mu.RUnlock()

	if max == -1 {
		m.ComputeMaxScrollOffset()
		m.mu.RLock()
		max = m.MaxOffset
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
	scroll := m.Scroll
	m.mu.RUnlock()

	if scroll == nil || scroll.Content == nil {
		return
	}

	contentHeight := scroll.Content.MinSize().Height
	viewportHeight := scroll.Size().Height

	m.mu.Lock()
	if contentHeight <= viewportHeight {
		m.MaxOffset = 0
	} else {
		m.MaxOffset = contentHeight - viewportHeight
	}
	m.mu.Unlock()
}

func (m *AppLayout) RefreshLayout() {
	JC.MainDebouncer.Call("refreshing_layout_layout", 5*time.Millisecond, func() {
		m.mu.RLock()
		if m.Scroll != nil {
			fyne.Do(m.Scroll.Refresh)
		}
		m.mu.RUnlock()
	})
}

func (m *AppLayout) RefreshContainer() {
	JC.MainDebouncer.Call("refreshing_layout_container", 5*time.Millisecond, func() {
		m.mu.RLock()
		fyne.Do(m.Container.Refresh)
		m.mu.RUnlock()
	})
}

func (m *AppLayout) Refresh() {
	if m == nil {
		return
	}

	m.mu.Lock()
	m.MaxOffset = -1
	currentState := m.state
	content := m.Content
	scroll := m.Scroll
	m.mu.Unlock()

	if content == nil || !AppStatusManager.IsReady() {
		m.mu.Lock()
		scroll.Content = m.Loading
		m.state = -1
		m.mu.Unlock()
	} else if !AppStatusManager.ValidConfig() {
		m.mu.Lock()
		scroll.Content = m.ActionFixSetting
		m.state = -2
		m.mu.Unlock()
	} else if !AppStatusManager.ValidCryptos() {
		m.mu.Lock()
		scroll.Content = m.ActionGetCryptos
		m.state = -3
		m.mu.Unlock()
	} else if !AppStatusManager.ValidPanels() {
		m.mu.Lock()
		scroll.Content = m.ActionAddPanel
		m.state = 0
		m.mu.Unlock()
	} else if !AppStatusManager.HasError() {
		m.mu.Lock()
		scroll.Content = *content
		m.state = AppStatusManager.panels_count
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		scroll.Content = m.Error
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
		tickers := m.Tickers
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

func NewAppLayoutManager() fyne.CanvasObject {

	JC.NotificationContainer = JW.NewNotificationDisplayWidget()
	DragPlaceholder = canvas.NewRectangle(JC.Transparent)
	if rect, ok := DragPlaceholder.(*canvas.Rectangle); ok {
		rect.CornerRadius = JC.PanelBorderRadius
	}

	manager := &AppLayout{
		TopBar: NewTopBar(),
	}

	AppLayoutManager = manager

	manager.Loading = NewAppPage(nil, "Loading...", nil)
	manager.Error = NewAppPage(nil, "Failed to start application...", nil)

	contentIcon := theme.ContentAddIcon()
	manager.ActionAddPanel = NewAppPage(&contentIcon, "Add Panel", func() {
		AppActionManager.CallButton("add_panel")
	})

	settingIcon := theme.SettingsIcon()
	manager.ActionFixSetting = NewAppPage(&settingIcon, "Open Settings", func() {
		AppActionManager.CallButton("open_settings")
	})

	restoreIcon := theme.ViewRestoreIcon()
	manager.ActionGetCryptos = NewAppPage(&restoreIcon, "Fetch Crypto Data", func() {
		AppActionManager.CallButton("refresh_cryptos")
	})

	// Create scroll container
	manager.Scroll = container.NewVScroll(nil)
	manager.Refresh()

	// Create tickers container
	manager.Tickers = container.NewWithoutLayout()

	// Tracking main container
	manager.Container = *container.New(
		&AppMainLayout{
			Padding: 10,
		},
		manager.TopBar,
		manager.Tickers,
		manager.Scroll,
		DragPlaceholder,
	)

	return fynetooltip.AddWindowToolTipLayer(
		&manager.Container,
		JC.Window.Canvas())
}
