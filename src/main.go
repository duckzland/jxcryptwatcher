package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"net/url"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"
)

var Grid *fyne.Container
var Window fyne.Window
var NotificationBox *widget.Label

// @todo Move these to theme
var appBG color.RGBA = color.RGBA{R: 13, G: 20, B: 33, A: 255}
var panelBG color.RGBA = color.RGBA{R: 50, G: 53, B: 70, A: 255}
var textColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var redColor color.RGBA = color.RGBA{R: 191, G: 8, B: 8, A: 255}
var greenColor color.RGBA = color.RGBA{R: 2, G: 115, B: 78, A: 255}

const epsilon = 1e-9

func main() {

	ExchangeCache.Reset()

	ConfigInit()

	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	// Don't invoke this before app.New(), binding.UntypedList will crash
	PanelsInit()

	Window = a.NewWindow("JXCrypto Watcher")

	Grid = container.New(NewDynamicGridWrapLayout(fyne.NewSize(300, 150)))

	list := BP.Get()
	for range list {
		Grid.Add(generateEmptyPanel())
	}

	NotificationBox = widget.NewLabel("")

	topBg := canvas.NewRectangle(panelBG)
	topBg.CornerRadius = 4
	topBg.SetMinSize(fyne.NewSize(860, 20))
	topBar := container.New(
		&stretchLayout{Widths: []float32{0.798, 0.004, 0.048, 0.002, 0.048, 0.002, 0.048, 0.002, 0.048}},
		container.NewStack(
			topBg,
			NotificationBox,
		),
		layout.NewSpacer(),

		// Reload cryptos.json
		NewHoverCursorIconButton("", theme.ViewRestoreIcon(), "Refresh ticker data", func() {
			doActionWithNotification("Fetching new ticker data...", "Finished fetching ticker data", NotificationBox, func() {
				Cryptos := CryptosType{}
				BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
				BP.InvalidatePanels()
				updateDisplay()
			})
		}),
		layout.NewSpacer(),

		// Refresh data from exchange
		NewHoverCursorIconButton("", theme.ViewRefreshIcon(), "Update rates from exchange", func() {
			doActionWithNotification("Fetching exchange rates...", "Panel refreshed with new rates", NotificationBox, func() {
				updateRates()
				updateDisplay()
			})
		}),
		layout.NewSpacer(),

		// Open settings form
		NewHoverCursorIconButton("", theme.SettingsIcon(), "Open settings", func() {
			generateSettingsForm()
		}),
		layout.NewSpacer(),

		// Add new panel
		NewHoverCursorIconButton("", theme.ContentAddIcon(), "Add new panel", func() {
			generatePanelForm("new")
		}),
	)

	bg := canvas.NewRectangle(appBG)
	bg.SetMinSize(fyne.NewSize(920, 600))

	Window.SetContent(fynetooltip.AddWindowToolTipLayer(container.NewStack(
		bg,
		container.NewPadded(
			container.NewBorder(
				topBar, nil, nil, nil, container.NewVScroll(Grid),
			),
		),
	), Window.Canvas()))

	Window.Resize(fyne.NewSize(920, 400))

	go func() {
		for {
			fyne.Do(func() {
				updateDisplay()
			})
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		for {
			doActionWithNotification("Fetching exchange rate...", "Fetching rates from exchange...", NotificationBox, func() {
				if updateRates() {
					fyne.Do(func() {
						updateDisplay()
					})
				}
			})

			time.Sleep(time.Duration(Config.Delay) * time.Second)
		}
	}()

	Window.ShowAndRun()

}

func generatePanelForm(panelKey string) {

	cm := BP.GetOptions()

	valueEntry := NewNumericalEntry(true)
	sourceEntry := NewCompletionEntry(cm)
	targetEntry := NewCompletionEntry(cm)
	decimalsEntry := NewNumericalEntry(false)

	title := "Adding New Panel"
	if panelKey == "new" {
		// Debug prefilled form
		// valueEntry.SetText("123")
		// sourceEntry.SetText("35626")
		// targetEntry.SetText("5426")
		// Default entry for decimals use 6 decimals!
		decimalsEntry.SetText("6")
	} else {
		pi := BP.GetIndex(panelKey)
		pkt := BP.GetDataByIndex(pi)

		title = "Editing Panel"

		valueEntry.SetDefaultValue(
			strconv.FormatFloat(pkt.GetSourceValueFloat(), 'f', NumDecPlaces(pkt.GetSourceValueFloat()), 64),
		)
		sourceEntry.SetDefaultValue(
			BP.GetDisplayById(pkt.GetSourceCoinString()),
		)
		targetEntry.SetDefaultValue(
			BP.GetDisplayById(pkt.GetTargetCoinString()),
		)
		decimalsEntry.SetDefaultValue(
			pkt.GetDecimalsString(),
		)
	}

	valueEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		value, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("Only numerical number with decimals allowed")
		}

		if math.Abs(value) < epsilon || value <= 0 {
			return fmt.Errorf("Only number larger than zero allowed")
		}

		return nil
	}

	sourceEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		tid := BP.GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !BP.ValidateId(id) {
			return fmt.Errorf("Invalid crypto selected")
		}

		xid := BP.GetIdByDisplay(targetEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err != nil && BP.ValidateId(bid) && bid == id {
			return fmt.Errorf("Cannot have the same coin for both source and target")
		}

		return nil
	}

	targetEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		tid := BP.GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !BP.ValidateId(id) {
			return fmt.Errorf("Invalid crypto selected")
		}

		xid := BP.GetIdByDisplay(targetEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err != nil && BP.ValidateId(bid) && bid == id {
			return fmt.Errorf("Cannot have the same coin for both source and target")
		}

		return nil
	}

	decimalsEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("Only numerical value without decimals allowed")
		}

		if x < 0 {
			return fmt.Errorf("Only number larger than zero allowed")
		}
		return nil
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Value", valueEntry),
		widget.NewFormItem("Source", sourceEntry),
		widget.NewFormItem("Target", targetEntry),
		widget.NewFormItem("Decimals", decimalsEntry),
	}

	d := NewExtendedFormDialog(title, formItems, func(b bool) {
		if b {
			if panelKey == "new" {
				ns := BP.Append(BP.GenerateKey(
					BP.GetIdByDisplay(sourceEntry.Text),
					BP.GetIdByDisplay(targetEntry.Text),
					valueEntry.Text,
					decimalsEntry.Text,
					0,
				))

				if ns != nil {
					Grid.Add(generatePanel(ns))
					ns.index = len(Grid.Objects)
				}

			} else {
				pi := BP.GetIndex(panelKey)
				if pi != -1 {
					npk := BP.GenerateKey(
						BP.GetIdByDisplay(sourceEntry.Text),
						BP.GetIdByDisplay(targetEntry.Text),
						valueEntry.Text,
						decimalsEntry.Text,
						0,
					)
					ns := BP.GetDataByIndex(pi)

					if ns != nil {
						opk := ns.oldKey

						// BugFix: Use set directly, don't use Update! to force update key
						ns.Set(npk)

						// Now we try to update, this can fail to set update!
						ns.Update(npk)

						nnpk := ns.Get()

						if BP.ValidatePanel(opk) && !BP.ValidatePanel(nnpk) {
							ns.index = -1
						}

						if !BP.ValidatePanel(opk) && BP.ValidatePanel(nnpk) {
							ns.index = -1
						}
					}
				}
			}

			doActionWithNotification("Saving Panel...", "Panel data saved...", NotificationBox, func() {
				if SavePanels() {
					fyne.Do(func() {
						Grid.Refresh()
					})

					if updateRates() {
						fyne.Do(func() {
							updateDisplay()
						})
					}
				}
			})
		}
	}, Window)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func generateSettingsForm() {

	delayEntry := NewNumericalEntry(false)
	dataEndPointEntry := widget.NewEntry()
	exchangeEndPointEntry := widget.NewEntry()

	delayEntry.SetDefaultValue(strconv.FormatInt(Config.Delay, 10))
	dataEndPointEntry.SetText(Config.DataEndpoint)
	exchangeEndPointEntry.SetText(Config.ExchangeEndpoint)

	delayEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("Only numerical value without decimals allowed")
		}

		if x < 0 {
			return fmt.Errorf("Only number larger than zero allowed")
		}

		return nil
	}

	dataEndPointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}

		return nil
	}

	exchangeEndPointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}

		return nil
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Ticker URL", dataEndPointEntry),
		widget.NewFormItem("Exchange URL", exchangeEndPointEntry),
		widget.NewFormItem("Delay(seconds)", delayEntry),
	}

	d := NewExtendedFormDialog("Settings", formItems, func(b bool) {
		if b {

			delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)

			Config.DataEndpoint = dataEndPointEntry.Text
			Config.ExchangeEndpoint = exchangeEndPointEntry.Text
			Config.Delay = delay

			doActionWithNotification("Saving configuration...", "Configuration data saved...", NotificationBox, func() {
				Config.SaveFile()
			})
		}
	}, Window)

	d.Show()
	d.Resize(fyne.NewSize(800, 300))
}

func generateEmptyPanel() fyne.CanvasObject {

	content := canvas.NewText("Loading...", textColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 16

	return panelItem(
		container.New(
			layout.NewCustomPaddedVBoxLayout(6),
			layout.NewSpacer(),
			content,
			layout.NewSpacer(),
		),
		panelBG,
		6,
		[4]float32{0, 5, 10, 5},
	)
}

func generatePanel(pdt *PanelDataType) fyne.CanvasObject {

	title := canvas.NewText(pdt.FormatTitle(), textColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	subtitle := canvas.NewText(pdt.FormatSubtitle(), textColor)
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = 16

	content := canvas.NewText(pdt.FormatContent(), textColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 30

	background := canvas.NewRectangle(panelBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = 6

	str := pdt.GetData()
	str.AddListener(binding.NewDataListener(func() {

		// No change needed
		if !pdt.DidChange() {
			return
		}

		ns := pdt.IsValueIncrease()

		if ns == 1 {
			background.FillColor = greenColor
			background.Refresh()
		}

		if ns == -1 {
			background.FillColor = redColor
			background.Refresh()
		}

		title.Text = pdt.FormatTitle()
		subtitle.Text = pdt.FormatSubtitle()
		content.Text = pdt.FormatContent()

		StartFlashingText(content, 50*time.Millisecond, textColor, 1)
	}))

	action := container.NewHBox(
		layout.NewSpacer(),
		NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
			fyne.Do(func() {
				dynpk, _ := str.Get()
				generatePanelForm(dynpk)
			})
		}),
		NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
			doActionWithNotification("Removing Panel...", "Panel removed...", NotificationBox, func() {

				dynpk, _ := str.Get()
				dynpi := BP.GetIndex(dynpk)

				RemovePanel(dynpi)
				SavePanels()
			})
		}),
	)

	return NewDoubleClickContainer(
		"ValidPanel",
		panelItem(
			container.NewStack(
				background,
				container.NewVBox(
					layout.NewSpacer(),
					title, content, subtitle,
					layout.NewSpacer(),
				),
				container.NewVBox(
					action,
				),
			),
			panelBG,
			6,
			[4]float32{0, 5, 10, 5},
		),
		action,
		false,
	)
}

func generateInvalidPanel(pk string) fyne.CanvasObject {

	pi := BP.GetIndex(pk)

	content := canvas.NewText("Invalid Panel", textColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 16

	action := container.NewHBox(
		layout.NewSpacer(),
		NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
			fyne.Do(func() {
				generatePanelForm(pk)
			})
		}),
		NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
			doActionWithNotification("Removing Panel...", "Panel removed...", NotificationBox, func() {

				RemovePanel(pi)

				SavePanels()
			})
		}),
	)

	return NewDoubleClickContainer(
		"InvalidPanel",
		panelItem(
			container.NewStack(
				container.NewVBox(
					layout.NewSpacer(),
					content,
					layout.NewSpacer(),
				),
				container.NewVBox(
					action,
				),
			),
			panelBG,
			6,
			[4]float32{0, 5, 10, 5},
		),
		action,
		false,
	)
}

func updateDisplay() {

	list := BP.Get()
	for i := range list {
		// Always get linked data! do not use the copied
		pkt := BP.GetDataByIndex(i)
		pk := pkt.Get()

		if BP.ValidatePanel(pk) {
			if pkt.Update(pk) {
				if pkt.index == -1 {
					npk := pkt.Get()
					// This panel hasnt been generated yet, create the markup!
					if BP.ValidatePanel(npk) {
						Grid.Objects[i] = generatePanel(pkt)
					} else {
						Grid.Objects[i] = generateInvalidPanel(pk)
					}
				}
			}
		} else {
			if pkt.index == -1 {
				Grid.Objects[i] = generateInvalidPanel(pk)
			}
		}

		pkt.index = i
	}

	// log.Print("Display Refreshed")
}

func updateRates() bool {

	// Clear cached rates
	ExchangeCache.Reset()
	ex := ExchangeDataType{}
	list := BP.Get()
	for i := range list {
		// Always get linked data! do not use the copied
		pkt := BP.GetDataByIndex(i)
		pk := pkt.Get()
		ex.GetRate(pk)
	}

	log.Print("Exchange Rate updated")

	return true
}

func panelItem(content fyne.CanvasObject, bgColor color.Color, borderRadius float32, padding [4]float32) fyne.CanvasObject {

	background := canvas.NewRectangle(bgColor)
	background.SetMinSize(fyne.NewSize(100, 100))

	if borderRadius != 0 {
		background.CornerRadius = borderRadius
	}

	item := container.NewStack(
		background,
		content,
	)

	// Simulate padding using empty spacers
	top := canvas.NewRectangle(color.Transparent)
	top.SetMinSize(fyne.NewSize(0, padding[0])) // top padding

	left := canvas.NewRectangle(color.Transparent)
	left.SetMinSize(fyne.NewSize(padding[1], 0)) // left padding

	bottom := canvas.NewRectangle(color.Transparent)
	bottom.SetMinSize(fyne.NewSize(0, padding[2])) // bottom padding

	right := canvas.NewRectangle(color.Transparent)
	right.SetMinSize(fyne.NewSize(padding[3], 0)) // right padding

	return container.NewBorder(top, bottom, left, right, item)
}

func doActionWithNotification(showText string, completeText string, box *widget.Label, callback func()) {

	go func() {

		callback()

		fyne.Do(func() {
			box.SetText(showText)
		})

		time.Sleep(3000 * time.Millisecond)

		fyne.Do(func() {
			box.SetText(completeText)
		})

		time.Sleep(8000 * time.Millisecond)

		fyne.Do(func() {
			box.SetText("")
		})
	}()
}

func StartFlashingText(text *canvas.Text, interval time.Duration, visibleColor color.Color, flashes int) {
	go func() {
		for i := 0; i < flashes*2; i++ { // 2 toggles per flash
			time.Sleep(interval)
			if i%2 == 0 {
				SetTextAlpha(text, 200)
			} else {
				SetTextAlpha(text, 255)
			}
			fyne.Do(func() {
				text.Refresh()
			})
		}
	}()
}

func SetTextAlpha(text *canvas.Text, alpha uint8) {
	switch c := text.Color.(type) {
	case color.RGBA:
		c.A = alpha
		text.Color = c
	case color.NRGBA:
		c.A = alpha
		text.Color = c
	default:
		// fallback to white with new alpha if type is unknown
		text.Color = color.RGBA{R: 255, G: 255, B: 255, A: alpha}
	}
}
