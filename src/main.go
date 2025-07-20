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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	fynetooltip "github.com/dweymouth/fyne-tooltip"
)

var Grid *fyne.Container
var Window fyne.Window
var NotificationBox *widget.Label

// @todo Move these to theme
var appBG color.RGBA = color.RGBA{R: 57, G: 62, B: 70, A: 255}
var panelBG color.RGBA = color.RGBA{R: 34, G: 40, B: 49, A: 255}
var textColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}

const epsilon = 1e-9

func main() {
	a := app.New()
	Window = a.NewWindow("JXCrypto Watcher")

	checkConfig()
	checkCryptos()

	// Don't invoke this before app.New(), binding.StringList will crash
	checkPanels()

	Grid = container.New(NewDynamicGridWrapLayout(fyne.NewSize(300, 150)))
	list, _ := BindedData.Get()
	for range list {
		Grid.Add(generateEmptyPanel())
	}

	NotificationBox = widget.NewLabel("")

	topBg := canvas.NewRectangle(panelBG)
	topBg.SetMinSize(fyne.NewSize(860, 20))
	topBar := container.New(
		&stretchLayout{Widths: []float32{0.80, 0.05, 0.05, 0.05, 0.05}},
		container.NewStack(
			topBg,
			NotificationBox,
		),
		// Reload cryptos.json
		NewHoverCursorIconButton("", theme.ViewRestoreIcon(), "Refresh ticker data", func() {
			doActionWithNotification("Fetching new ticker data...", "Finished fetching ticker data", NotificationBox, func() {
				refreshCryptos()
			})
		}),

		// Refresh data from exchange
		NewHoverCursorIconButton("", theme.ViewRefreshIcon(), "Update rates from exchange", func() {
			doActionWithNotification("Fetching exchange rates...", "Panel refreshed with new rates", NotificationBox, func() {
				updateData(true)
			})
		}),

		// Open settings form
		NewHoverCursorIconButton("", theme.SettingsIcon(), "Open settings", func() {
			generateSettingsForm()
		}),

		// Add new panel
		NewHoverCursorIconButton("", theme.ContentAddIcon(), "Add new panel", func() {
			generatePanelForm("new")
		}),
	)

	bg := canvas.NewRectangle(appBG)
	bg.SetMinSize(fyne.NewSize(920, 400))

	Window.SetContent(fynetooltip.AddWindowToolTipLayer(container.NewStack(
		bg,
		container.NewPadded(
			container.NewVBox(
				topBar,
				Grid,
			),
		),
	), Window.Canvas()))

	Window.Resize(fyne.NewSize(920, 400))

	go func() {
		for {
			doActionWithNotification("Fetching exchange rate...", "Updating panel...", NotificationBox, func() {
				updateData(true)
			})

			time.Sleep(time.Duration(Config.Delay) * time.Second)
		}
	}()

	Window.ShowAndRun()

}

func generatePanelForm(panelKey string) {

	cm := getTickerOptions()

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
		// decimalsEntry.SetText("6")
	} else {
		title = "Editing Panel"
		source := getTickerDisplayById(strconv.FormatInt(getPanelSourceCoin(panelKey), 10))
		target := getTickerDisplayById(strconv.FormatInt(getPanelTargetCoin(panelKey), 10))
		value := strconv.FormatFloat(getPanelSourceValue(panelKey), 'f', NumDecPlaces(getPanelSourceValue(panelKey)), 64)
		decimals := strconv.FormatInt(getPanelDecimals(panelKey), 10)

		valueEntry.SetDefaultValue(value)
		sourceEntry.SetDefaultValue(source)
		targetEntry.SetDefaultValue(target)
		decimalsEntry.SetDefaultValue(decimals)
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

		tid := getTickerIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !validateCryptoId(id) {
			return fmt.Errorf("Invalid crypto selected")
		}
		return nil
	}

	targetEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		tid := getTickerIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !validateCryptoId(id) {
			return fmt.Errorf("Invalid crypto selected")
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

			source, _ := strconv.ParseInt(getTickerIdByDisplay(sourceEntry.Text), 10, 64)
			target, _ := strconv.ParseInt(getTickerIdByDisplay(targetEntry.Text), 10, 64)
			value, _ := strconv.ParseFloat(valueEntry.Text, 64)
			decimals, _ := strconv.ParseInt(decimalsEntry.Text, 10, 64)

			if panelKey == "new" {
				appendPanel(generatePanelKey(PanelType{
					Source:   source,
					Target:   target,
					Value:    value,
					Decimals: decimals,
				}, 0))

			} else {
				pi := getPanel(panelKey)

				if pi != -1 {
					insertPanel(generatePanelKey(PanelType{
						Source:   source,
						Target:   target,
						Value:    value,
						Decimals: decimals,
					}, 0), pi)
				}
			}

			doActionWithNotification("Saving Panel...", "Panel data saved...", NotificationBox, func() {
				saved := savePanels()
				if saved {
					fyne.Do(func() {
						Grid.Refresh()
					})
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
				saveConfig()
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

func generatePanel(pk string) fyne.CanvasObject {

	pi := getPanel(pk)
	p := message.NewPrinter(language.English)

	decimals := getPanelDecimals(pk)
	sourceValue := getPanelSourceValue(pk)
	targetValue := getPanelValue(pk)

	sourceCoin := getPanelSourceCoin(pk)
	targetCoin := getPanelTargetCoin(pk)

	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := getTickerSymbolById(sourceID)

	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := getTickerSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	evt := p.Sprintf("%v", number.Decimal(targetValue, number.MaxFractionDigits(int(decimals))))
	sts := p.Sprintf("%v", number.Decimal(sourceValue, number.MaxFractionDigits(frac)))
	tts := p.Sprintf("%v", number.Decimal(sourceValue*float64(targetValue), number.MaxFractionDigits(frac)))

	// Debug
	// tts := fmt.Sprintf(ttd, panel.Value*data.TargetAmount+(rand.Float64()*5))

	title := canvas.NewText(fmt.Sprintf("%s %s to %s", sts, sourceSymbol, targetSymbol), textColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	subtitle := canvas.NewText(fmt.Sprintf("%s %s = %s %s", "1", sourceSymbol, evt, targetSymbol), textColor)
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = 16

	content := canvas.NewText(fmt.Sprintf("%s %s", tts, targetSymbol), textColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 30

	action := container.NewHBox(
		layout.NewSpacer(),
		NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
			fyne.Do(func() {
				generatePanelForm(pk)
			})
		}),
		NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
			doActionWithNotification("Removing Panel...", "Panel removed...", NotificationBox, func() {
				// Async
				removePanel(pi)
				saved := savePanels()
				if saved {
					fyne.Do(func() {
						Grid.Refresh()
					})
				}
			})
		}),
	)
	return NewDoubleClickContainer(
		panelItem(
			container.NewStack(
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

	pi := getPanel(pk)

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
				// Async
				removePanel(pi)
				saved := savePanels()
				if saved {
					fyne.Do(func() {
						Grid.Refresh()
					})
				}
			})
		}),
	)

	return NewDoubleClickContainer(
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

func updateData(isAsync bool) bool {

	updated := false

	list, _ := BindedData.Get()
	for i, val := range list {

		if validatePanel(val) {
			updated = updatePanel(val)
		} else {
			Grid.Objects[i] = generateInvalidPanel(val)
			updated = true
		}
	}

	if updated {
		// Must refresh via grid, refreshing via individual panel or only relying on databind change will not work!
		if isAsync {
			fyne.Do(func() {
				Grid.Refresh()
			})
		} else {
			Grid.Refresh()
		}
	}

	log.Print("Rate updated")

	return updated
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

		time.Sleep(10000 * time.Millisecond)

		fyne.Do(func() {
			box.SetText("")
		})
	}()
}
