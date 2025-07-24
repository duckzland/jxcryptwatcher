package widgets

import (
	"fmt"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

func NewSettingsForm(onSave func()) *ExtendedFormDialog {

	delayEntry := NewNumericalEntry(false)
	dataEndPointEntry := widget.NewEntry()
	exchangeEndPointEntry := widget.NewEntry()

	// Prefill with config data
	delayEntry.SetDefaultValue(strconv.FormatInt(JT.Config.Delay, 10))
	dataEndPointEntry.SetText(JT.Config.DataEndpoint)
	exchangeEndPointEntry.SetText(JT.Config.ExchangeEndpoint)

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
		widget.NewFormItem("Delay (seconds)", delayEntry),
	}

	return NewExtendedFormDialog("Settings", formItems, func(b bool) {
		if b {
			delay, _ := strconv.ParseInt(delayEntry.Text, 10, 64)
			JT.Config.DataEndpoint = dataEndPointEntry.Text
			JT.Config.ExchangeEndpoint = exchangeEndPointEntry.Text
			JT.Config.Delay = delay

			DoActionWithNotification("Saving configuration...", "Configuration data saved...", JC.NotificationBox, func() {
				if onSave != nil {
					onSave()
				}
			})
		}
	}, JC.Window)
}
