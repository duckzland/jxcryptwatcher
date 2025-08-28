package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
)

type AppMainLayout struct {
	Padding float32
}

var AppLayoutManager *AppLayout = nil

func (a *AppMainLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	// Background setup
	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(size)
	bg.Resize(size)

	topBar := objects[0]
	content := objects[1]

	topHeight := topBar.MinSize().Height
	contentY := topHeight + 2*a.Padding
	newContentWidth := size.Width - 2*a.Padding
	newContentHeight := size.Height - contentY - 2*a.Padding

	// Update global layout dimensions
	JC.MainLayoutContentWidth = newContentWidth
	JC.MainLayoutContentHeight = newContentHeight

	// TopBar layout
	newTopBarPos := fyne.NewPos(a.Padding, a.Padding)
	newTopBarSize := fyne.NewSize(newContentWidth, topHeight)

	if topBar.Position() != newTopBarPos || topBar.Size() != newTopBarSize {
		topBar.Move(newTopBarPos)
		topBar.Resize(newTopBarSize)
		topBar.Refresh()
	}

	// Content layout
	newContentPos := fyne.NewPos(a.Padding, contentY)
	newContentSize := fyne.NewSize(newContentWidth, newContentHeight)

	if content.Position() != newContentPos || content.Size() != newContentSize {
		content.Move(newContentPos)
		content.Resize(newContentSize)
		content.Refresh()
	}
}

func (a *AppMainLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	top := objects[0].MinSize()
	content := objects[1].MinSize()

	width := fyne.Max(top.Width, content.Width) + 2*a.Padding
	height := top.Height + content.Height + 3*a.Padding

	return fyne.NewSize(width, height)
}

type AppLayout struct {
	TopBar     *fyne.CanvasObject
	Content    *fyne.Container
	Scroll     *container.Scroll
	AddButton  *EmptyClickablePanel
	Loading    *LoadingPanel
	OnNewPanel func()
}

func (m *AppLayout) SetContent(container *fyne.Container) {
	m.Content = container
}

func (m *AppLayout) Refresh() {

	if m.Content == nil {
		m.Scroll.Content = m.Loading
	} else if len(m.Content.Objects) == 0 {
		m.Scroll.Content = m.AddButton
	} else {
		m.Scroll.Content = m.Content
	}

	fyne.Do(func() {
		m.Scroll.Refresh()
	})
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

func NewAppLayoutManager(topbar *fyne.CanvasObject, content *fyne.Container, onNewPanel func()) fyne.CanvasObject {
	manager := &AppLayout{
		TopBar:     topbar,
		Content:    content,
		OnNewPanel: onNewPanel,
	}

	manager.Loading = NewLoadingPanel()
	manager.AddButton = NewEmptyClickablePanel(func() {
		if manager.OnNewPanel != nil {
			manager.OnNewPanel()
		}
	})

	// Create scroll container
	manager.Scroll = container.NewVScroll(nil)
	manager.Refresh()

	AppLayoutManager = manager

	return fynetooltip.AddWindowToolTipLayer(
		container.New(
			&AppMainLayout{
				Padding: 10,
			},
			*manager.TopBar,
			manager.Scroll,
		),
		JC.Window.Canvas())
}
