package panels

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewPanelDisplay(
	pdt *JT.PanelDataType,
	onEdit func(pk string, uuid string),
	onDelete func(uuid string),
) fyne.CanvasObject {

	// Generate a new UUID for the panel, avoiding panel use wrong uuid
	uuid := JC.CreateUUID()
	pdt.ID = uuid

	title := canvas.NewText("", JC.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = JC.PanelTitleSize

	subtitle := canvas.NewText("", JC.TextColor)
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = JC.PanelSubTitleSize

	content := canvas.NewText("", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = JC.PanelContentSize

	background := canvas.NewRectangle(JC.PanelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.PanelBorderRadius

	str := pdt.GetData()

	action := NewPanelActionBar(
		func() {
			dynpk, _ := str.Get()
			// Potential bug fix, maybe this will persist and store as local variable?
			u := uuid
			if onEdit != nil {
				log.Printf("Editing panel %s", u)
				onEdit(dynpk, u)
			}
		},
		func() {
			// Potential bug fix, maybe this will persist and store as local variable?
			u := uuid
			if onDelete != nil {
				log.Printf("Deleting panel %s", u)
				onDelete(u)
			}
		},
	)

	panel := JW.NewDoubleClickContainer(
		uuid,
		NewPanelContainer(
			container.NewStack(
				background,
				container.NewVBox(
					layout.NewSpacer(),
					title, content, subtitle,
					layout.NewSpacer(),
				),
				container.NewVBox(action),
			),
		),
		action,
		false,
	)

	str.AddListener(binding.NewDataListener(func() {
		if !pdt.DidChange() {
			return
		}

		switch pdt.IsValueIncrease() {
		case 1:
			background.FillColor = JC.GreenColor
			background.Refresh()
		case -1:
			background.FillColor = JC.RedColor
			background.Refresh()
		}

		updateContent(pdt, title, subtitle, content, background, panel)

		JW.StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	updateContent(pdt, title, subtitle, content, background, panel)

	return panel
}

func updateContent(pdt *JT.PanelDataType, title, subtitle, content *canvas.Text, background *canvas.Rectangle, panel *JW.DoubleClickContainer) {

	// Invalid panel
	if !JT.BP.ValidatePanel(pdt.Get()) {
		title.Text = "Invalid Panel"
		subtitle.Hide()
		content.Hide()
		background.FillColor = JC.PanelBG

		return
	}

	// Fresh panel
	if pdt.UsePanelKey().GetValueFloat() == -1 {
		title.Text = "Loading..."
		subtitle.Hide()
		content.Hide()
		panel.DisableClick()
		background.FillColor = JC.PanelBG

		return
	}

	// Normal panel
	title.Text = pdt.FormatTitle()
	subtitle.Text = pdt.FormatSubtitle()
	content.Text = pdt.FormatContent()

	subtitle.Show()
	content.Show()
	panel.EnableClick()
}
