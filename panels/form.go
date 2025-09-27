package panels

import (
	"fmt"
	"math"
	"math/big"
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
	onSave func(pdt JT.PanelData),
	onNew func(pdt JT.PanelData),
	onRender func(layer *fyne.Container),
	onDestroy func(layer *fyne.Container),
) JW.DialogForm {

	cm := JT.UsePanelMaps().GetOptions()
	cs := JT.UsePanelMaps().GetMaps().GetSearchMap()

	pse := container.NewStack()
	pte := container.NewStack()
	pop := []*fyne.Container{pse, pte}

	valueEntry := JW.NewNumericalEntry(true)
	sourceEntry := JW.NewCompletionEntry(cm, cs, pse)
	targetEntry := JW.NewCompletionEntry(cm, cs, pte)
	decimalsEntry := JW.NewNumericalEntry(false)

	title := "Adding New Panel"

	if panelKey != "new" {

		pkt := JT.UsePanelMaps().GetDataByID(uuid)
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

		sourceEntry.SetDefaultValue(JT.UsePanelMaps().GetDisplayById(pko.GetSourceCoinString()))

		targetEntry.SetDefaultValue(JT.UsePanelMaps().GetDisplayById(pko.GetTargetCoinString()))

		decimalsEntry.SetDefaultValue(pko.GetDecimalsString())

	} else {
		decimalsEntry.SetText("6")
	}

	valueEntry.SetValidator(func(s string) error {

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
	})

	sourceEntry.SetValidator(func(s string) error {

		if len(s) == 0 {
			return fmt.Errorf("Please select a cryptocurrency.")
		}

		tid := JT.UsePanelMaps().GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.UsePanelMaps().ValidateId(id) {
			return fmt.Errorf("Please select a valid cryptocurrency.")
		}

		xid := JT.UsePanelMaps().GetIdByDisplay(targetEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err == nil && JT.UsePanelMaps().ValidateId(bid) && bid == id {
			return fmt.Errorf("Source and target cryptocurrencies must be different.")
		}

		return nil
	})

	targetEntry.SetValidator(func(s string) error {

		if len(s) == 0 {
			return fmt.Errorf("Please select a cryptocurrency.")
		}

		tid := JT.UsePanelMaps().GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.UsePanelMaps().ValidateId(id) {
			return fmt.Errorf("Please select a valid cryptocurrency.")
		}

		xid := JT.UsePanelMaps().GetIdByDisplay(sourceEntry.Text)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err == nil && JT.UsePanelMaps().ValidateId(bid) && bid == id {
			return fmt.Errorf("Source and target cryptocurrencies must be different.")
		}

		return nil
	})

	decimalsEntry.SetValidator(func(s string) error {
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
	})

	formItems := []*widget.FormItem{
		widget.NewFormItem("Value", valueEntry),
		widget.NewFormItem("Source", sourceEntry),
		widget.NewFormItem("Target", targetEntry),
		widget.NewFormItem("Decimals", decimalsEntry),
	}

	parent := JW.NewDialogForm(title, formItems, nil, nil, pop,
		func(b bool) bool {
			if b {
				npk := JT.NewPanelKey()
				var ns JT.PanelData

				sid := JT.UsePanelMaps().GetIdByDisplay(sourceEntry.Text)
				tid := JT.UsePanelMaps().GetIdByDisplay(targetEntry.Text)
				bid := JT.UsePanelMaps().GetSymbolById(sid)
				mid := JT.UsePanelMaps().GetSymbolById(tid)

				newKey := npk.GenerateKey(
					sid,
					tid,
					valueEntry.Text,
					bid,
					mid,
					decimalsEntry.Text,
					big.NewFloat(-1),
				)

				if panelKey == "new" {
					ns = JT.UsePanelMaps().Append(newKey)

					if ns == nil {
						JC.Notify("Unable to add new panel. Please try again.")
						return false
					}

					ns.SetStatus(JC.STATE_FETCHING_NEW)

					if onNew != nil {
						onNew(ns)
					}

				} else {
					ns = JT.UsePanelMaps().GetDataByID(uuid)
					if ns == nil {
						JC.Notify("Unable to update panel. Please try again.")
						return false
					}

					pkt := ns.UsePanelKey()
					nkt := JT.NewPanelKey()
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

			return true
		},
		onRender,
		onDestroy,
		JC.Window)

	sourceEntry.SetParent(parent)
	targetEntry.SetParent(parent)

	return parent
}
