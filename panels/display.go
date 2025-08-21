package panels

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

type PanelLayout struct{}

func (p *PanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 5 {
		return
	}

	bg := objects[0]
	title := objects[1]
	content := objects[2]
	subtitle := objects[3]
	action := objects[4]

	// Full-size background
	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	// Filter visible content elements
	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{title, content, subtitle} {
		if obj.Visible() && obj.MinSize().Height > 0 {
			centerItems = append(centerItems, obj)
		}
	}

	// Compute total height of visible center items
	var totalHeight float32
	for _, obj := range centerItems {
		totalHeight += obj.MinSize().Height
	}

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	// Center stack
	for _, obj := range centerItems {
		objSize := obj.MinSize()
		obj.Move(fyne.NewPos((size.Width-objSize.Width)/2, currentY))
		obj.Resize(objSize)
		currentY += objSize.Height
	}

	// Position action element last (highest z-index)
	actionSize := action.MinSize()
	action.Move(fyne.NewPos(size.Width-actionSize.Width, 0))
	action.Resize(actionSize)
}

func (p *PanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	for _, obj := range objects[1:4] {
		if obj.Visible() && obj.MinSize().Height > 0 {
			size := obj.MinSize()
			if size.Width > width {
				width = size.Width
			}
			height += size.Height
		}
	}

	return fyne.NewSize(width, height)
}

var panel_group *string

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
				JC.Logf("Editing panel %s", u)
				go onEdit(dynpk, u)
			}
		},
		func() {
			// Potential bug fix, maybe this will persist and store as local variable?
			u := uuid
			if onDelete != nil {
				JC.Logf("Deleting panel %s", u)
				go onDelete(u)
			}
		},
	)

	pg := "panels"
	panel_group = &pg

	panel := JW.NewDoubleClickContainer(
		uuid,
		container.New(&PanelLayout{},
			background,
			title,
			content,
			subtitle,
			action),
		action,
		false,
		panel_group,
	)

	str.AddListener(binding.NewDataListener(func() {

		// BugFix: Use pdt status to check the didchange logic preventing panel never change after rates updated
		if !pdt.DidChange() && pdt.Status == 1 {
			JC.Logln("Skipping panel panel rebuild for:", pdt.OldKey, pdt.Get(), pdt.Status)
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

		JA.StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	updateContent(pdt, title, subtitle, content, background, panel)

	return panel
}

func updateContent(pdt *JT.PanelDataType, title, subtitle, content *canvas.Text, background *canvas.Rectangle, panel *JW.DoubleClickContainer) {

	// Mutate from fresh panel to normal panel when we got valid value
	if pdt.UsePanelKey().GetValueFloat() != -1 {
		pdt.Status = 1
	}

	// New Panel
	if pdt.Status == -1 {
		title.Text = "Fetching Rates..."
		subtitle.Hide()
		content.Hide()
		panel.DisableClick()
		background.FillColor = JC.PanelBG

		return
	}

	// Invalid panel
	if !JT.BP.ValidatePanel(pdt.Get()) {
		title.Text = "Invalid Panel"
		subtitle.Hide()
		content.Hide()
		background.FillColor = JC.PanelBG

		return
	}

	// Fresh panel
	if pdt.Status == 0 {
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
