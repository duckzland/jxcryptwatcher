package panels

import (
	"fmt"
	"math"
	"strconv"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewPanelForm(
	panelKey string,
	uuid string,
	onSave func(),
	onNew func(pdt *JT.PanelDataType),
) *JW.ExtendedFormDialog {

	cm := JT.BP.GetOptions()

	popupTarget := container.NewStack()

	valueEntry := JW.NewNumericalEntry(true)
	sourceEntry := JW.NewCompletionEntry(cm, popupTarget)
	targetEntry := JW.NewCompletionEntry(cm, popupTarget)
	decimalsEntry := JW.NewNumericalEntry(false)

	title := "Adding New Panel"
	if panelKey != "new" {

		pkt := JT.BP.GetData(uuid)
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

	parent := JW.NewExtendedFormDialog(title, formItems, nil, popupTarget, func(b bool) {
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
					JC.Notify("Unable to add new panel. Please try again.")
					return
				}
				if onNew != nil {
					onNew(ns)
				}
				ns.Status = -1
			} else {
				ns := JT.BP.GetData(uuid)
				if ns == nil {
					JC.Notify("Unable to update panel. Please try again.")
					return
				}
				ns.Set(newKey)
				ns.Update(newKey)
			}

			if onSave != nil {
				onSave()
			}
		}
	}, JC.Window)

	sourceEntry.Parent = parent
	targetEntry.Parent = parent

	return parent
}
