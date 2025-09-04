package apps

import (
	"fmt"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewSettingsForm(onSave func()) *JW.ExtendedFormDialog {

	delayEntry := JW.NewNumericalEntry(false)
	dataEndPointEntry := widget.NewEntry()
	exchangeEndPointEntry := widget.NewEntry()

	tickerCMC100EndpointEntry := widget.NewEntry()
	tickerFearGreedEndpointEntry := widget.NewEntry()
	tickerMetricsEndpointEntry := widget.NewEntry()
	tickerListingsEndpointEntry := widget.NewEntry()
	tickerDelayEntry := JW.NewNumericalEntry(false)
	proApiKeyEntry := widget.NewEntry()

	// Prefill with config data
	delayEntry.SetDefaultValue(strconv.FormatInt(JT.Config.Delay, 10))
	dataEndPointEntry.SetText(JT.Config.DataEndpoint)
	exchangeEndPointEntry.SetText(JT.Config.ExchangeEndpoint)

	tickerCMC100EndpointEntry.SetText(JT.Config.TickerCMC100Endpoint)
	tickerFearGreedEndpointEntry.SetText(JT.Config.TickerFearGreedEndpoint)
	tickerMetricsEndpointEntry.SetText(JT.Config.TickerMetricsEndpoint)
	tickerListingsEndpointEntry.SetText(JT.Config.TickerListingsEndpoint)
	tickerDelayEntry.SetDefaultValue(strconv.FormatInt(JT.Config.TickerDelay, 10))
	proApiKeyEntry.SetText(JT.Config.ProApiKey)

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

	tickerCMC100EndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	tickerFearGreedEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	tickerMetricsEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	tickerListingsEndpointEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		_, err := url.ParseRequestURI(s)
		if err != nil {
			return fmt.Errorf("Invalid URL format")
		}
		return nil
	}

	tickerDelayEntry.Validator = func(s string) error {
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
		widget.NewFormItem("Crypto Maps URL", dataEndPointEntry),
		widget.NewFormItem("Exchange URL", exchangeEndPointEntry),
		widget.NewFormItem("Delay (seconds)", delayEntry),
		widget.NewFormItem("CMC100 URL", tickerCMC100EndpointEntry),
		widget.NewFormItem("Fear&Greed URL", tickerFearGreedEndpointEntry),
		widget.NewFormItem("Metrics URL", tickerMetricsEndpointEntry),
		widget.NewFormItem("Listings URL", tickerListingsEndpointEntry),
		widget.NewFormItem("Ticker (seconds)", tickerDelayEntry),
		widget.NewFormItem("CMC Pro API Key", proApiKeyEntry),
	}

	return JW.NewExtendedFormDialog("Settings", formItems, nil, nil, func(b bool) {
		if b {

			delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)
			tickerDelay, _ := strconv.ParseInt(tickerDelayEntry.Text, 10, 64)

			JT.Config.DataEndpoint = dataEndPointEntry.Text
			JT.Config.ExchangeEndpoint = exchangeEndPointEntry.Text
			JT.Config.Delay = delay

			JT.Config.TickerCMC100Endpoint = tickerCMC100EndpointEntry.Text
			JT.Config.TickerFearGreedEndpoint = tickerFearGreedEndpointEntry.Text
			JT.Config.TickerMetricsEndpoint = tickerMetricsEndpointEntry.Text
			JT.Config.TickerListingsEndpoint = tickerListingsEndpointEntry.Text
			JT.Config.TickerDelay = tickerDelay
			JT.Config.ProApiKey = proApiKeyEntry.Text

			if onSave != nil {
				onSave()
			}
		}
	}, JC.Window)
}
