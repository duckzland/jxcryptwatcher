package main

import (
	"fmt"
	"image/color"
	"math"
	"net/url"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Grid *fyne.Container
var BindedData binding.StringList
var Window fyne.Window

// @todo Move these to theme
var appBG color.RGBA = color.RGBA{R: 57, G: 62, B: 70, A: 255}
var panelBG color.RGBA = color.RGBA{R: 34, G: 40, B: 49, A: 255}
var textColor color.RGBA = color.RGBA{R: 255, G: 255, B: 255, A: 255}

const epsilon = 1e-9

/**
 * Main function
 */
func main() {

	checkConfig()
	checkCryptos()
	checkPanels()

	a := app.New()
	Window = a.NewWindow("JXCrypto Watcher")

	BindedData = binding.NewStringList()

	Grid = container.New(NewDynamicGridWrapLayout(fyne.NewSize(300, 150)))

	for _, panel := range Panels {
		BindedData.Append(generatePanelKey(panel, 0))
		generateEmptyPanel()
	}

	topBg := canvas.NewRectangle(panelBG)
	topBg.SetMinSize(fyne.NewSize(860, 20))
	topBar := container.New(
		&stretchLayout{Widths: []float32{0.70, 0.15, 0.15}},
		topBg,
		widget.NewButton("Settings", func() {
			generateSettingsForm()
		}),
		widget.NewButton("Add Panel", func() {
			generatePanelForm("new")
		}),
	)

	bg := canvas.NewRectangle(appBG)
	bg.SetMinSize(fyne.NewSize(920, 400))

	Window.SetContent(container.NewStack(
		bg,
		container.NewPadded(
			container.NewVBox(
				topBar,
				Grid,
			),
		),
	))

	Window.Resize(fyne.NewSize(920, 400))

	go func() {
		for {
			fyne.DoAndWait(func() {
				updateData()
			})

			time.Sleep(time.Duration(Config.Delay) * time.Second)
		}
	}()

	Window.ShowAndRun()

}

/**
 * Helper function for generating panel config form
 */
func generatePanelForm(panelKey string) {

	valueEntry := NewNumericalEntry(true)
	sourceEntry := NewCompletionEntry(CryptosOptions)
	targetEntry := NewCompletionEntry(CryptosOptions)
	decimalsEntry := NewNumericalEntry(false)

	if panelKey == "new" {
		// Debug prefilled form
		// valueEntry.SetText("123")
		// sourceEntry.SetText("35626")
		// targetEntry.SetText("5426")
		// decimalsEntry.SetText("6")
	} else {
		pi := getPanelByKey(panelKey)
		if pi != -1 {
			panel := Panels[pi]
			source := getTickerDisplayById(strconv.FormatInt(panel.Source, 10))
			target := getTickerDisplayById(strconv.FormatInt(panel.Target, 10))
			value := strconv.FormatFloat(panel.Value, 'f', NumDecPlaces(panel.Value), 64)
			decimals := strconv.FormatInt(panel.Decimals, 10)

			valueEntry.SetDefaultValue(value)
			sourceEntry.SetDefaultValue(source)
			targetEntry.SetDefaultValue(target)
			decimalsEntry.SetDefaultValue(decimals)
		}
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

	d := dialog.NewForm("New Panel Entry", "Save", "Cancel", formItems, func(b bool) {
		if b {

			source, _ := strconv.ParseInt(getTickerIdByDisplay(sourceEntry.Text), 10, 64)
			target, _ := strconv.ParseInt(getTickerIdByDisplay(targetEntry.Text), 10, 64)
			value, _ := strconv.ParseFloat(valueEntry.Text, 64)
			decimals, _ := strconv.ParseInt(decimalsEntry.Text, 10, 64)

			if panelKey == "new" {
				appendPanel(PanelType{
					Source:   source,
					Target:   target,
					Value:    value,
					Decimals: decimals,
				})

			} else {
				pi := getPanelByKey(panelKey)

				if pi != -1 {
					insertPanel(PanelType{
						Source:   source,
						Target:   target,
						Value:    value,
						Decimals: decimals,
					}, pi)
				}
			}

			Grid.Refresh()
			savePanels()
		}
	}, Window)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

/**
 * Helper function for generating settings form
 */
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

	d := dialog.NewForm("Settings", "Save", "Cancel", formItems, func(b bool) {
		if b {

			delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)

			Config.DataEndpoint = dataEndPointEntry.Text
			Config.ExchangeEndpoint = exchangeEndPointEntry.Text
			Config.Delay = delay

			saveConfig()
		}
	}, Window)

	d.Show()
	d.Resize(fyne.NewSize(800, 300))
}

/**
 * Helper function for generating empty panel
 */
func generateEmptyPanel() {

	content := canvas.NewText("Loading...", textColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 16

	Grid.Add(panelItem(
		container.New(
			layout.NewCustomPaddedVBoxLayout(6),
			layout.NewSpacer(),
			content,
			layout.NewSpacer(),
		),
		panelBG,
		6,
		[4]float32{0, 5, 10, 5},
	))
}

/**
 * Helper for generate a single panel
 */
func generatePanel(panelKey string, index int) {
	pi := getPanelByKey(panelKey)

	if pi != -1 && len(Panels) > pi {
		panel := Panels[pi]
		data := getExchangeData(panel)

		tvs := fmt.Sprintf("%s%d%s", "%.", panel.Decimals, "f")
		evt := fmt.Sprintf(tvs, data.TargetAmount)

		std := fmt.Sprintf("%s%d%s", "%.", NumDecPlaces(panel.Value), "f")
		sts := fmt.Sprintf(std, panel.Value)

		// ttd := fmt.Sprintf("%s%d%s", "%.", NumDecPlaces(panel.Value), "f")
		ttd := fmt.Sprintf("%s%d%s", "%.", 2, "f")
		tts := fmt.Sprintf(ttd, panel.Value*data.TargetAmount)

		// Debug
		// tts := fmt.Sprintf(ttd, panel.Value*data.TargetAmount+(rand.Float64()*5))

		title := canvas.NewText(fmt.Sprintf("%s %s to %s", sts, data.SourceSymbol, data.TargetSymbol), textColor)
		title.Alignment = fyne.TextAlignCenter
		title.TextStyle = fyne.TextStyle{Bold: true}
		title.TextSize = 16

		subtitle := canvas.NewText(fmt.Sprintf("%s %s = %s %s", "1", data.SourceSymbol, evt, data.TargetSymbol), textColor)
		subtitle.Alignment = fyne.TextAlignCenter
		subtitle.TextSize = 16

		content := canvas.NewText(fmt.Sprintf("%s %s", tts, data.TargetSymbol), textColor)
		content.Alignment = fyne.TextAlignCenter
		content.TextStyle = fyne.TextStyle{Bold: true}
		content.TextSize = 30

		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			removePanel(pi)
			savePanels()
		})

		editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
			generatePanelForm(panelKey)
		})

		action := container.NewHBox(
			layout.NewSpacer(),
			editBtn,
			deleteBtn,
		)

		// This is hack to stretch the width
		rect := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 0})
		rect.SetMinSize(fyne.NewSize(270, 120))

		panelDisplay := panelItem(
			NewDoubleClickContainer(
				container.NewStack(
					rect,
					container.NewVBox(
						layout.NewSpacer(),
						title, content, subtitle,
						layout.NewSpacer(),
					),
					container.NewVBox(
						action,
					),
				),
				action,
			),
			panelBG,
			6,
			[4]float32{0, 5, 10, 5},
		)

		if index == -1 {
			Grid.Add(panelDisplay)
		} else if len(Grid.Objects) > index {
			Grid.Objects[index] = panelDisplay
		}
	}
}

/**
 * Helper for update the panels
 */
func updateData() bool {

	updated := false

	list, _ := BindedData.Get()
	for i, val := range list {
		pi := getPanelByKey(val)

		if pi != -1 && len(Panels) > pi {
			panel := Panels[pi]
			data := getExchangeData(panel)
			tv := generatePanelKey(panel, float32(data.TargetAmount))

			if tv != val {
				updatePanel(panel, pi, tv)
				updated = true
			}

		} else {
			removePanel(i)
			updated = true
		}
	}

	if updated {
		// Must refresh via grid, refreshing via individual panel or only relying on databind change will not work!
		Grid.Refresh()
	}

	return updated
}

/**
 * Helper function for decorating panel item with background, border radius and padding
 */
func panelItem(content fyne.CanvasObject, bgColor color.Color, borderRadius float32, padding [4]float32) fyne.CanvasObject {

	background := canvas.NewRectangle(bgColor)
	background.SetMinSize(fyne.NewSize(100, 100))

	if borderRadius != 0 {
		background.CornerRadius = borderRadius
	}

	item := container.NewStack(
		background,
		container.NewCenter(content),
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

/**
 * Helper function for removing entry by its index
 */
func removeAt(index int, list binding.StringList) {
	values, _ := list.Get()
	if index < 0 || index >= len(values) {
		return // avoid out-of-bounds
	}

	// Remove item at index
	updated := append(values[:index], values[index+1:]...)
	list.Set(updated)
}
