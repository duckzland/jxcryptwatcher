package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

func NewPanel(pdt *JT.PanelDataType, onEdit func(pk string), onDelete func(index int)) fyne.CanvasObject {
	title := canvas.NewText(pdt.FormatTitle(), JC.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	subtitle := canvas.NewText(pdt.FormatSubtitle(), JC.TextColor)
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = 16

	content := canvas.NewText(pdt.FormatContent(), JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 30

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = 6

	str := pdt.GetData()
	str.AddListener(binding.NewDataListener(func() {
		if !pdt.DidChange() {
			return
		}

		switch pdt.IsValueIncrease() {
		case 1:
			background.FillColor = JC.GreenColor
		case -1:
			background.FillColor = JC.RedColor
		}
		background.Refresh()

		title.Text = pdt.FormatTitle()
		subtitle.Text = pdt.FormatSubtitle()
		content.Text = pdt.FormatContent()

		StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	action := NewPanelActionBar(
		func() {
			dynpk, _ := str.Get()
			if onEdit != nil {
				onEdit(dynpk)
			}
		},
		func() {
			dynpk, _ := str.Get()
			dynpi := JT.BP.GetIndex(dynpk)

			if onEdit != nil {
				onDelete(dynpi)
			}
		},
	)

	return NewDoubleClickContainer(
		"ValidPanel",
		NewPanelItem(
			container.NewStack(
				background,
				container.NewVBox(
					layout.NewSpacer(),
					title, content, subtitle,
					layout.NewSpacer(),
				),
				container.NewVBox(action),
			),
			JC.PanelBG,
			6,
			[4]float32{0, 5, 10, 5},
		),
		action,
		false,
	)
}
