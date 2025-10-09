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
	onSave func(pdt JT.PanelData),
	onNew func(pdt JT.PanelData),
	onRender func(layer *fyne.Container),
	onDestroy func(layer *fyne.Container),
) JW.DialogForm {

	var allowValidation bool = false

	validateValue := func(s string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("This field is required")
		}
		value, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("No decimals allowed")
		}
		if math.Abs(value) < JC.EPSILON || value <= 0 {
			return fmt.Errorf("Must larger than zero")
		}
		return nil
	}

	validateCoin := func(s string, other string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("Please select a cryptocurrency")
		}
		tid := JT.UsePanelMaps().GetIdByDisplay(s)
		id, err := strconv.ParseInt(tid, 10, 64)
		if err != nil || !JT.UsePanelMaps().ValidateId(id) {
			return fmt.Errorf("Invalid cryptocurrency selected")
		}
		xid := JT.UsePanelMaps().GetIdByDisplay(other)
		bid, err := strconv.ParseInt(xid, 10, 64)
		if err == nil && JT.UsePanelMaps().ValidateId(bid) && bid == id {
			return fmt.Errorf("Source and target must different")
		}
		return nil
	}

	validateDecimals := func(s string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("This field cannot be empty")
		}
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("No decimals allowed")
		}
		if x < 0 {
			return fmt.Errorf("Must larger than zero")
		}
		if x > 20 {
			return fmt.Errorf("Maximum 20 decimal digits")
		}
		return nil
	}

	cm := JT.UsePanelMaps().GetOptions()
	cs := JT.UsePanelMaps().GetMaps().GetSearchMap()

	pse := container.NewStack()
	pte := container.NewStack()
	pop := []*fyne.Container{pse, pte}

	ve := JW.NewNumericalEntry(true)
	se := JW.NewCompletionEntry(cm, cs, pse)
	te := JW.NewCompletionEntry(cm, cs, pte)
	de := JW.NewNumericalEntry(false)

	title := "Adding New Panel"

	if panelKey != "new" {

		pkt := JT.UsePanelMaps().GetDataByID(uuid)
		pko := pkt.UsePanelKey()

		title = "Editing Panel"

		ve.SetDefaultValue(
			strconv.FormatFloat(
				pko.GetSourceValueFloat(),
				'f',
				JC.NumDecPlaces(pko.GetSourceValueFloat()),
				64,
			),
		)

		se.SetDefaultValue(JT.UsePanelMaps().GetDisplayById(pko.GetSourceCoinString()))

		te.SetDefaultValue(JT.UsePanelMaps().GetDisplayById(pko.GetTargetCoinString()))

		de.SetDefaultValue(pko.GetDecimalsString())

	} else {
		de.SetText("6")
	}

	ve.Validator = validateValue
	se.Validator = func(s string) error {
		return validateCoin(s, te.Text)
	}
	te.Validator = func(s string) error {
		return validateCoin(s, se.Text)
	}
	de.Validator = validateDecimals

	fi := []*widget.FormItem{
		widget.NewFormItem("Source Amount", ve),
		widget.NewFormItem("From Cryptocurrency", se),
		widget.NewFormItem("To Cryptocurrency", te),
		widget.NewFormItem("Decimal Precision", de),
	}

	parent := JW.NewDialogForm(title, fi, nil, nil, pop,
		func() bool {
			defer func() { allowValidation = false }()
			allowValidation = true

			hasError := false

			if ve.Validate() != nil {
				hasError = true
			}
			if se.Validate() != nil {
				hasError = true
			}
			if te.Validate() != nil {
				hasError = true
			}
			if de.Validate() != nil {
				hasError = true
			}

			if hasError {
				return false
			}

			npk := JT.NewPanelKey()
			var ns JT.PanelData

			sid := JT.UsePanelMaps().GetIdByDisplay(se.Text)
			tid := JT.UsePanelMaps().GetIdByDisplay(te.Text)
			bid := JT.UsePanelMaps().GetSymbolById(sid)
			mid := JT.UsePanelMaps().GetSymbolById(tid)

			npk.Set(npk.GenerateKey(
				sid,
				tid,
				ve.Text,
				bid,
				mid,
				de.Text,
				JC.ToBigFloat(-1),
			))

			if panelKey == "new" {
				ns = JT.UsePanelMaps().Append(npk.GetRawValue())

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

				if pkt.GetSourceCoinInt() != npk.GetSourceCoinInt() || pkt.GetTargetCoinInt() != npk.GetTargetCoinInt() {
					ns.SetStatus(JC.STATE_LOADING)
					ns.Set(npk.GetRawValue())
					ns.Update(npk.GetRawValue())
				} else {
					opk := ns.GetOldKey()
					npk.UpdateValue(pkt.GetValueFloat())
					ns.Set(npk.GetRawValue())
					ns.SetOldKey(opk)
				}
			}

			if onSave != nil {
				onSave(ns)
			}

			return true
		},
		onRender,
		onDestroy,
		JC.Window)

	se.SetParent(parent)
	te.SetParent(parent)

	return parent
}
