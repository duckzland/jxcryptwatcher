package panels

import (
	"fmt"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewPanelForm(
	panelKey string,
	uuid string,
	onSave func(pdt *JT.PanelDataType),
	onNew func(pdt *JT.PanelDataType),
) *JW.ExtendedFormDialog {

	cm := JT.BP.GetOptions()

	popupSourceEntryTarget := container.NewStack()
	popupTargetEntryTarget := container.NewStack()

	valueEntry := JW.NewNumericalEntry(true)
	sourceEntry := JW.NewCompletionEntry(cm, popupSourceEntryTarget)
	targetEntry := JW.NewCompletionEntry(cm, popupTargetEntryTarget)
	decimalsEntry := JW.NewNumericalEntry(false)

	title := "Adding New Panel"
	if panelKey != "new" {

		pkt := JT.BP.GetDataByID(uuid)
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

		if math.Abs(value) < JC.EPSILON || value <= 0 {
			return fmt.Errorf("Only number larger than zero allowed")
		}

		return nil
	}

	sourceEntry.SetValidator(func(s string) error {

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
	})

	targetEntry.SetValidator(func(s string) error {

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
	})

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

		if x > 20 {
			return fmt.Errorf("Maximum supported precision is 20 decimal digits")
		}

		return nil
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Value", valueEntry),
		widget.NewFormItem("Source", sourceEntry),
		widget.NewFormItem("Target", targetEntry),
		widget.NewFormItem("Decimals", decimalsEntry),
	}

	popupTarget := []*fyne.Container{popupSourceEntryTarget, popupTargetEntryTarget}

	parent := JW.NewExtendedFormDialog(title, formItems, nil, popupTarget, func(b bool) {
		if b {
			var npk JT.PanelKeyType
			var ns *JT.PanelDataType
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
				ns = JT.BP.Append(newKey)

				if ns == nil {
					JC.Notify("Unable to add new panel. Please try again.")
					return
				}

				ns.SetStatus(JC.STATE_FETCHING_NEW)

				if onNew != nil {
					onNew(ns)
				}

			} else {
				ns = JT.BP.GetDataByID(uuid)
				if ns == nil {
					JC.Notify("Unable to update panel. Please try again.")
					return
				}

				pkt := ns.UsePanelKey()
				nkt := JT.PanelKeyType{}
				nkt.Set(newKey)

				// Coin change, need to refresh data and invalidate the rates
				if pkt.GetSourceCoinInt() != nkt.GetSourceCoinInt() || pkt.GetTargetCoinInt() != nkt.GetTargetCoinInt() {
					ns.SetStatus(JC.STATE_LOADING)
					ns.Set(newKey)
				}

				ns.Update(newKey)
			}

			if onSave != nil {
				onSave(ns)
			}
		}
	}, JC.Window)

	sourceEntry.SetParent(parent)
	targetEntry.SetParent(parent)

	return parent
}
