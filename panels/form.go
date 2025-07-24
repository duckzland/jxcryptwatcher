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
		title = "Editing Panel"

		valueEntry.SetDefaultValue(strconv.FormatFloat(pkt.UsePanelKey().GetSourceValueFloat(), 'f',
			JC.NumDecPlaces(pkt.UsePanelKey().GetSourceValueFloat()), 64))

		sourceEntry.SetDefaultValue(JT.BP.GetDisplayById(pkt.UsePanelKey().GetSourceCoinString()))

		targetEntry.SetDefaultValue(JT.BP.GetDisplayById(pkt.UsePanelKey().GetTargetCoinString()))

		decimalsEntry.SetDefaultValue(pkt.UsePanelKey().GetDecimalsString())

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
			return fmt.Errorf("This field cannot be empty")
		}

		tid := JT.BP.GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.BP.ValidateId(id) {
			return fmt.Errorf("Invalid crypto selected")
		}

		xid := JT.BP.GetIdByDisplay(targetEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err != nil && JT.BP.ValidateId(bid) && bid == id {
			return fmt.Errorf("Cannot have the same coin for both source and target")
		}

		return nil
	}

	targetEntry.Validator = sourceEntry.Validator

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
				0,
			)

			if panelKey == "new" {
				ns := JT.BP.Append(newKey)
				if ns != nil {
					if onNew != nil {
						onNew(ns)
					}
					ns.Index = len(JC.Grid.Objects)
				}
			} else {
				pi := JT.BP.GetIndex(panelKey)
				if pi != -1 {
					ns := JT.BP.GetDataByIndex(pi)
					if ns != nil {
						opk := ns.OldKey
						ns.Set(newKey)
						ns.Update(newKey)
						nnpk := ns.Get()

						if JT.BP.ValidatePanel(opk) && !JT.BP.ValidatePanel(nnpk) {
							ns.Index = -1
						}
						if !JT.BP.ValidatePanel(opk) && JT.BP.ValidatePanel(nnpk) {
							ns.Index = -1
						}
					}
				}
			}

			JW.DoActionWithNotification("Saving Panel...", "Panel data saved...", JC.NotificationBox, func() {
				if onSave != nil {
					onSave()
				}
			})
		}
	}, JC.Window)
}
