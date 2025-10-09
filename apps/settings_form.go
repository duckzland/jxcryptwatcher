package apps

import (
	"errors"
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

	var allowValidation bool = false
	validateURL := func(s string) error {
		if !allowValidation {
			return nil
		}
		if s == "" {
			return errors.New("This field is required")
		}
		u, err := url.ParseRequestURI(s)
		if err != nil {
			return errors.New("Invalid URL format")
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("Only http or https allowed")
		}

		return nil
	}

	validateDelay := func(s string) error {
		if !allowValidation {
			return nil
		}
		if s == "" {
			return errors.New("This field is required")
		}
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.New("No decimals allowed")
		}
		if val < 0 {
			return errors.New("Must larger than zero")
		}
		return nil
	}

	delay := JW.NewNumericalEntry(false)
	cryptos := JW.NewTextEntry()
	exchange := JW.NewTextEntry()
	altindex := JW.NewTextEntry()
	feargreed := JW.NewTextEntry()
	cmc100 := JW.NewTextEntry()
	marketcap := JW.NewTextEntry()
	rsi := JW.NewTextEntry()
	etf := JW.NewTextEntry()
	dominance := JW.NewTextEntry()

	delay.SetDefaultValue(strconv.FormatInt(JT.UseConfig().Delay, 10))
	cryptos.SetText(JT.UseConfig().DataEndpoint)
	exchange.SetText(JT.UseConfig().ExchangeEndpoint)
	altindex.SetText(JT.UseConfig().AltSeasonEndpoint)
	feargreed.SetText(JT.UseConfig().FearGreedEndpoint)
	cmc100.SetText(JT.UseConfig().CMC100Endpoint)
	marketcap.SetText(JT.UseConfig().MarketCapEndpoint)
	rsi.SetText(JT.UseConfig().RSIEndpoint)
	etf.SetText(JT.UseConfig().ETFEndpoint)
	dominance.SetText(JT.UseConfig().DominanceEndpoint)

	delay.Validator = validateDelay
	cryptos.Validator = validateURL
	exchange.Validator = validateURL
	altindex.Validator = validateURL
	feargreed.Validator = validateURL
	cmc100.Validator = validateURL
	marketcap.Validator = validateURL
	rsi.Validator = validateURL
	etf.Validator = validateURL
	dominance.Validator = validateURL

	items := []*widget.FormItem{
		widget.NewFormItem("Crypto Maps URL", cryptos),
		widget.NewFormItem("Exchange URL", exchange),
		widget.NewFormItem("AltSeason URL", altindex),
		widget.NewFormItem("Fear&Greed URL", feargreed),
		widget.NewFormItem("CMC100 URL", cmc100),
		widget.NewFormItem("MarketCap URL", marketcap),
		widget.NewFormItem("RSI URL", rsi),
		widget.NewFormItem("ETF URL", etf),
		widget.NewFormItem("Dominance URL", dominance),
		widget.NewFormItem("Delay (seconds)", delay),
	}

	return JW.NewDialogForm("Settings", items, nil, nil, nil,
		func() bool {
			defer func() { allowValidation = false }()

			allowValidation = true
			hasError := false

			if cryptos.Validate() != nil {
				hasError = true
			}
			if exchange.Validate() != nil {
				hasError = true
			}
			if altindex.Validate() != nil {
				hasError = true
			}
			if feargreed.Validate() != nil {
				hasError = true
			}
			if cmc100.Validate() != nil {
				hasError = true
			}
			if marketcap.Validate() != nil {
				hasError = true
			}
			if rsi.Validate() != nil {
				hasError = true
			}
			if etf.Validate() != nil {
				hasError = true
			}
			if dominance.Validate() != nil {
				hasError = true
			}
			if delay.Validate() != nil {
				hasError = true
			}

			if hasError {
				return false
			}

			val, _ := strconv.ParseInt(delay.Text, 10, 64)
			JT.UseConfig().Delay = val
			JT.UseConfig().DataEndpoint = cryptos.Text
			JT.UseConfig().ExchangeEndpoint = exchange.Text
			JT.UseConfig().AltSeasonEndpoint = altindex.Text
			JT.UseConfig().FearGreedEndpoint = feargreed.Text
			JT.UseConfig().CMC100Endpoint = cmc100.Text
			JT.UseConfig().MarketCapEndpoint = marketcap.Text
			JT.UseConfig().RSIEndpoint = rsi.Text
			JT.UseConfig().ETFEndpoint = etf.Text
			JT.UseConfig().DominanceEndpoint = dominance.Text

			if onSave != nil {
				onSave()
			}

			return true
		},
		onRender,
		onDestroy,
		JC.Window)
}
