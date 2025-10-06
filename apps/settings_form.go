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
) JW.DialogForm {

	delayEntry := JW.NewNumericalEntry(false)
	dataEndPointEntry := widget.NewEntry()
	exchangeEndPointEntry := widget.NewEntry()
	altSeasonsEndpointEntry := widget.NewEntry()
	fearGreedEndpointEntry := widget.NewEntry()
	CMC100EndpointEntry := widget.NewEntry()
	marketCapEndpointEntry := widget.NewEntry()
	rsiEndpointEntry := widget.NewEntry()
	etfEndpointEntry := widget.NewEntry()
	dominanceEndpointEntry := widget.NewEntry()

	// Prefill with config data
	delayEntry.SetDefaultValue(strconv.FormatInt(JT.UseConfig().Delay, 10))
	dataEndPointEntry.SetText(JT.UseConfig().DataEndpoint)
	exchangeEndPointEntry.SetText(JT.UseConfig().ExchangeEndpoint)
	altSeasonsEndpointEntry.SetText(JT.UseConfig().AltSeasonEndpoint)
	fearGreedEndpointEntry.SetText(JT.UseConfig().FearGreedEndpoint)
	CMC100EndpointEntry.SetText(JT.UseConfig().CMC100Endpoint)
	marketCapEndpointEntry.SetText(JT.UseConfig().MarketCapEndpoint)
	rsiEndpointEntry.SetText(JT.UseConfig().RSIEndpoint)
	etfEndpointEntry.SetText(JT.UseConfig().ETFEndpoint)
	dominanceEndpointEntry.SetText(JT.UseConfig().DominanceEndpoint)

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

	rsiEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	etfEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	dominanceEndpointEntry.Validator = func(s string) error {
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
		widget.NewFormItem("RSI URL", rsiEndpointEntry),
		widget.NewFormItem("ETF URL", etfEndpointEntry),
		widget.NewFormItem("Dominance URL", dominanceEndpointEntry),
		widget.NewFormItem("Delay (seconds)", delayEntry),
	}

	return JW.NewDialogForm("Settings", formItems, nil, nil, nil,
		func(b bool) bool {
			if b {

				delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)

				JT.UseConfig().DataEndpoint = dataEndPointEntry.Text
				JT.UseConfig().ExchangeEndpoint = exchangeEndPointEntry.Text
				JT.UseConfig().AltSeasonEndpoint = altSeasonsEndpointEntry.Text
				JT.UseConfig().FearGreedEndpoint = fearGreedEndpointEntry.Text
				JT.UseConfig().CMC100Endpoint = CMC100EndpointEntry.Text
				JT.UseConfig().MarketCapEndpoint = marketCapEndpointEntry.Text
				JT.UseConfig().RSIEndpoint = rsiEndpointEntry.Text
				JT.UseConfig().ETFEndpoint = etfEndpointEntry.Text
				JT.UseConfig().DominanceEndpoint = dominanceEndpointEntry.Text

				JT.UseConfig().Delay = delay

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
