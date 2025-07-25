package panels

import (
	"fmt"
	"math"
	"strconv"

	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewPanelForm(
	panelKey string,
	onSave func(),
	onNew func(pdt *JT.PanelDataType),
) *JW.ExtendedFormDialog {

	cm := JT.BP.GetOptions()

	valueEntry := JW.NewNumericalEntry(true)
	sourceEntry := JW.NewCompletionEntry(cm)
	targetEntry := JW.NewCompletionEntry(cm)
	decimalsEntry := JW.NewNumericalEntry(false)

	title := "Adding New Panel"
	if panelKey != "new" {

		pi := JT.BP.GetIndex(panelKey)
		pkt := JT.BP.GetDataByIndex(pi)
		pko := pkt.UsePanelKey()

		title = "Editing Panel"

		valueEntry.SetDefaultValue(
			strconv.FormatFloat(
				pko.GetSourceValueFloat(),
				'f',
				JC.NumDecPlaces(pko.GetSourceValueFloat()),
				64,
			),
		)

		sourceEntry.SetDefaultValue(JT.BP.GetDisplayById(pko.GetSourceCoinString()))

		targetEntry.SetDefaultValue(JT.BP.GetDisplayById(pko.GetTargetCoinString()))

		decimalsEntry.SetDefaultValue(pko.GetDecimalsString())

	} else {
		decimalsEntry.SetText("6")
	}

	valueEntry.Validator = func(s string) error {

		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}

		value, err := strconv.ParseFloat(s, 64)

		if err != nil {
			return fmt.Errorf("Only numerical number with decimals allowed")
		}

		if math.Abs(value) < JC.Epsilon || value <= 0 {
			return fmt.Errorf("Only number larger than zero allowed")
		}

		return nil
	}

	sourceEntry.Validator = func(s string) error {

		if len(s) == 0 {
			return fmt.Errorf("Please select a cryptocurrency.")
		}

		tid := JT.BP.GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.BP.ValidateId(id) {
			return fmt.Errorf("Please select a valid cryptocurrency.")
		}

		xid := JT.BP.GetIdByDisplay(targetEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err == nil && JT.BP.ValidateId(bid) && bid == id {
			return fmt.Errorf("Source and target cryptocurrencies must be different.")
		}

		return nil
	}

	targetEntry.Validator = func(s string) error {

		if len(s) == 0 {
			return fmt.Errorf("Please select a cryptocurrency.")
		}

		tid := JT.BP.GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.BP.ValidateId(id) {
			return fmt.Errorf("Please select a valid cryptocurrency.")
		}

		xid := JT.BP.GetIdByDisplay(sourceEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err == nil && JT.BP.ValidateId(bid) && bid == id {
			return fmt.Errorf("Source and target cryptocurrencies must be different.")
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

	return JW.NewExtendedFormDialog(title, formItems, func(b bool) {
		if b {
			var npk JT.PanelKeyType
			newKey := npk.GenerateKey(
				JT.BP.GetIdByDisplay(sourceEntry.Text),
				JT.BP.GetIdByDisplay(targetEntry.Text),
				valueEntry.Text,
				JT.BP.GetSymbolByDisplay(sourceEntry.Text),
				JT.BP.GetSymbolByDisplay(targetEntry.Text),
				decimalsEntry.Text,
				-1,
			)

			if panelKey == "new" {
				ns := JT.BP.Append(newKey)
				if ns == nil {
					JW.DoActionWithNotification("Error", "Failed to add new panel.", JC.NotificationBox, nil)
					return
				}
				if onNew != nil {
					onNew(ns)
				}
				ns.Index = -1
			} else {
				pi := JT.BP.GetIndex(panelKey)
				if pi == -1 {
					JW.DoActionWithNotification("Error", "Panel not found for update.", JC.NotificationBox, nil)
					return
				}
				ns := JT.BP.GetDataByIndex(pi)
				if ns == nil {
					JW.DoActionWithNotification("Error", "Failed to update panel.", JC.NotificationBox, nil)
					return
				}
				ns.Set(newKey)
				ns.Update(newKey)
			}

			JW.DoActionWithNotification("Saving Panel...", "Panel data saved...", JC.NotificationBox, func() {
				if onSave != nil {
					onSave()
				}
			})
		}
	}, JC.Window)
}
