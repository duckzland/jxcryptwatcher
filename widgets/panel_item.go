package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
)

func NewPanelItem(content fyne.CanvasObject) fyne.CanvasObject {

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	item := container.NewStack(background, content)
	padding := JC.PanelPadding

	// Simulate padding using transparent spacers
	top := canvas.NewRectangle(color.Transparent)
	top.SetMinSize(fyne.NewSize(0, padding[0]))

	left := canvas.NewRectangle(color.Transparent)
	left.SetMinSize(fyne.NewSize(padding[1], 0))

	bottom := canvas.NewRectangle(color.Transparent)
	bottom.SetMinSize(fyne.NewSize(0, padding[2]))

	right := canvas.NewRectangle(color.Transparent)
	right.SetMinSize(fyne.NewSize(padding[3], 0))

	return container.NewBorder(top, bottom, left, right, item)
}
