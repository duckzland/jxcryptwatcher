package apps

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
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
	content := objects[2]

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
	TopBar           *fyne.CanvasObject
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

func (m *AppLayout) SetPage(container fyne.CanvasObject) {
	m.Content = &container
}

func (m *AppLayout) SetTickers(container *fyne.Container) {
	if container == nil {
		return
	}
	m.Tickers = container
	m.Container.Objects[1] = container
	m.RefreshContainer()
}

func (m *AppLayout) Refresh() {
	if m == nil {
		return
	}

	m.MaxOffset = -1
	currentState := m.state

	if m.Content == nil || !AppStatusManager.IsReady() {
		m.Scroll.Content = m.Loading
		m.state = -1
	} else if !AppStatusManager.ValidConfig() {
		m.Scroll.Content = m.ActionFixSetting
		m.state = -2
	} else if !AppStatusManager.ValidCryptos() {
		m.Scroll.Content = m.ActionGetCryptos
		m.state = -3
	} else if !AppStatusManager.ValidPanels() {
		m.Scroll.Content = m.ActionAddPanel
		m.state = 0
	} else if !AppStatusManager.HasError() {
		m.Scroll.Content = *m.Content
		m.state = AppStatusManager.panels_count
	} else {
		m.Scroll.Content = m.Error
		m.state = -5
	}

	if m.state != currentState {
		m.RefreshLayout()
	}

	if AppStatusManager.IsReady() {

		if !AppStatusManager.IsValidProKey() {
			m.SetTickers(container.NewCenter(
				widget.NewLabelWithStyle("Please update your CMC Pro API Key in Settings.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			))
			return
		}

		if m.Tickers != nil && m.Tickers != JC.Tickers && AppStatusManager.bad_tickers == false {
			m.SetTickers(JC.Tickers)
			return
		}

		if m.Tickers != nil && m.Tickers == JC.Tickers && AppStatusManager.bad_tickers == true {
			m.SetTickers(container.NewWithoutLayout())
			return
		}

		if m.Tickers == nil && AppStatusManager.bad_tickers == false {
			m.SetTickers(JC.Tickers)
			return
		}
	}

}

func (m *AppLayout) OffsetY() float32 {
	return m.Scroll.Offset.Y
}

func (m *AppLayout) OffsetX() float32 {
	return m.Scroll.Offset.X
}

func (m *AppLayout) Height() float32 {
	if m.Scroll == nil {
		return -1
	}
	return m.Scroll.Size().Height
}

func (m *AppLayout) Width() float32 {
	if m.Scroll == nil {
		return -1
	}

	return m.Scroll.Size().Width
}

func (m *AppLayout) IsReady() bool {
	return m.Scroll != nil && m.Content != nil
}

func (m *AppLayout) ScrollBy(delta float32) {
	current := m.OffsetY()
	newOffset := current + delta

	if m.MaxOffset == -1 {
		m.ComputeMaxScrollOffset()
	}

	if newOffset < 0 {
		if current > 0 {
			newOffset = 0
		} else {
			return
		}
	} else if newOffset > m.MaxOffset {
		if current < m.MaxOffset {
			newOffset = m.MaxOffset
		} else {
			return
		}
	}

	m.SetOffsetY(newOffset)
	m.RefreshLayout()
}

func (m *AppLayout) ComputeMaxScrollOffset() {

	if m.Scroll == nil || m.Scroll.Content == nil {
		return
	}

	contentHeight := m.Scroll.Content.MinSize().Height
	viewportHeight := m.Scroll.Size().Height

	if contentHeight <= viewportHeight {
		m.MaxOffset = 0
	} else {
		m.MaxOffset = contentHeight - viewportHeight
	}
}

func (m *AppLayout) SetOffsetY(offset float32) {
	if m.Scroll == nil || m.Scroll.Offset.Y == offset {
		return
	}

	m.Scroll.Offset.Y = offset
	m.RefreshLayout()
}

func (m *AppLayout) RefreshLayout() {
	JC.MainDebouncer.Call("refreshing_layout_layout", 5*time.Millisecond, func() {
		if m.Scroll != nil {
			fyne.Do(m.Scroll.Refresh)
		}
	})
}

func (m *AppLayout) RefreshContainer() {
	JC.MainDebouncer.Call("refreshing_layout_container", 5*time.Millisecond, func() {
		fyne.Do(m.Container.Refresh)
		JC.Logln("== Refreshing Container")
	})
}

func NewAppLayoutManager(topbar *fyne.CanvasObject) fyne.CanvasObject {

	DragPlaceholder = canvas.NewRectangle(JC.Transparent)
	if rect, ok := DragPlaceholder.(*canvas.Rectangle); ok {
		rect.CornerRadius = JC.PanelBorderRadius
	}

	manager := &AppLayout{
		TopBar: topbar,
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
		*manager.TopBar,
		manager.Tickers,
		manager.Scroll,
		DragPlaceholder,
	)

	return fynetooltip.AddWindowToolTipLayer(
		&manager.Container,
		JC.Window.Canvas())
}
