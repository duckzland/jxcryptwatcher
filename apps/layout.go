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

var AppMainPanelScrollWindow *container.Scroll = nil

func (a *AppMainLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	// Build background
	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(fyne.NewSize(920, 600))
	bg.Resize(size)

	topBar := objects[0]
	content := objects[1]

	// TopBar layout
	topHeight := topBar.MinSize().Height
	topBar.Move(fyne.NewPos(a.Padding, a.Padding))
	topBar.Resize(fyne.NewSize(size.Width-2*a.Padding, topHeight))

	// Content layout (scrollable)
	contentY := topHeight + 2*a.Padding
	JC.MainLayoutContentWidth = size.Width - 2*a.Padding
	JC.MainLayoutContentHeight = size.Height - contentY - 2*a.Padding

	content.Move(fyne.NewPos(a.Padding, contentY))
	content.Resize(fyne.NewSize(JC.MainLayoutContentWidth, JC.MainLayoutContentHeight))
}

func (a *AppMainLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	top := objects[0].MinSize()
	content := objects[1].MinSize()

	width := fyne.Max(top.Width, content.Width) + 2*a.Padding
	height := top.Height + content.Height + 3*a.Padding

	return fyne.NewSize(width, height)
}

func NewAppLayout(topbar *fyne.CanvasObject, content *fyne.Container) fyne.CanvasObject {
	AppMainPanelScrollWindow = container.NewVScroll(content)

	return fynetooltip.AddWindowToolTipLayer(
		container.New(
			&AppMainLayout{
				Padding: 10,
			},
			*topbar,
			AppMainPanelScrollWindow,
		),
		JC.Window.Canvas())
}
