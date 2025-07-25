package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
)

func NewAppLayout(bg *canvas.Rectangle, topbar *fyne.CanvasObject, content *fyne.Container) fyne.CanvasObject {
	return fynetooltip.AddWindowToolTipLayer(
		container.NewStack(
			bg,
			container.NewPadded(
				container.NewBorder(
					*topbar,
					nil, nil, nil,
					container.NewVScroll(content),
				),
			),
		), JC.Window.Canvas(),
	)
}
