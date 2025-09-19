package apps

import (
	"fmt"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewSettingsForm(
	onSave func(),
	onRender func(layer *fyne.Container),
	onDestroy func(layer *fyne.Container),
) *JW.DialogForm {

	delayEntry := JW.NewNumericalEntry(false)
	dataEndPointEntry := widget.NewEntry()
	exchangeEndPointEntry := widget.NewEntry()
	altSeasonsEndpointEntry := widget.NewEntry()
	fearGreedEndpointEntry := widget.NewEntry()
	CMC100EndpointEntry := widget.NewEntry()
	marketCapEndpointEntry := widget.NewEntry()

	// Prefill with config data
	delayEntry.SetDefaultValue(strconv.FormatInt(JT.Config.Delay, 10))
	dataEndPointEntry.SetText(JT.Config.DataEndpoint)
	exchangeEndPointEntry.SetText(JT.Config.ExchangeEndpoint)
	altSeasonsEndpointEntry.SetText(JT.Config.AltSeasonEndpoint)
	fearGreedEndpointEntry.SetText(JT.Config.FearGreedEndpoint)
	CMC100EndpointEntry.SetText(JT.Config.CMC100Endpoint)
	marketCapEndpointEntry.SetText(JT.Config.MarketCapEndpoint)

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

	altSeasonsEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	fearGreedEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	CMC100EndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	marketCapEndpointEntry.Validator = func(s string) error {
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
		widget.NewFormItem("Crypto Maps URL", dataEndPointEntry),
		widget.NewFormItem("Exchange URL", exchangeEndPointEntry),
		widget.NewFormItem("AltSeason URL", altSeasonsEndpointEntry),
		widget.NewFormItem("Fear&Greed URL", fearGreedEndpointEntry),
		widget.NewFormItem("CMC100 URL", CMC100EndpointEntry),
		widget.NewFormItem("MarketCap URL", marketCapEndpointEntry),
		widget.NewFormItem("Delay (seconds)", delayEntry),
	}

	return JW.NewDialogForm("Settings", formItems, nil, nil, nil,
		func(b bool) bool {
			if b {

				delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)

				JT.Config.DataEndpoint = dataEndPointEntry.Text
				JT.Config.ExchangeEndpoint = exchangeEndPointEntry.Text
				JT.Config.AltSeasonEndpoint = altSeasonsEndpointEntry.Text
				JT.Config.FearGreedEndpoint = fearGreedEndpointEntry.Text
				JT.Config.CMC100Endpoint = CMC100EndpointEntry.Text
				JT.Config.MarketCapEndpoint = marketCapEndpointEntry.Text

				JT.Config.Delay = delay

				if onSave != nil {
					onSave()
				}
			}

			return true
		},
		onRender,
		onDestroy,
		JC.Window)
}
