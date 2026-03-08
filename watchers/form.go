package watchers

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewWatcherForm(
	uuid string,
	onSave func(pdt JT.PanelData),
	onRender func(layer *fyne.Container),
	onDestroy func(layer *fyne.Container),
) JW.DialogForm {

	JC.PrintPerfStats("Opening watcher form", time.Now())

	var allowValidation bool = false

	validateOps := func(key int, s string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("This field is required")
		}
		if key > 2 || key < 0 {
			return fmt.Errorf("Invalid operation mode")
		}
		return nil
	}

	validateInt := func(s string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("This field is required")
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Must be an integer")
		}
		if val <= 0 {
			return fmt.Errorf("Must be greater than 0")
		}
		return nil
	}

	validateFloat := func(s string) error {
		if !allowValidation {
			return nil
		}
		if len(s) == 0 {
			return fmt.Errorf("This field is required")
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("Must be a number")
		}
		if val <= 0 {
			return fmt.Errorf("Must be greater than 0")
		}
		return nil
	}

	re := JW.NewNumericalEntry(true)  // Rate
	le := JW.NewNumericalEntry(false) // Limit
	de := JW.NewNumericalEntry(false) // Duration

	oe := JW.NewSelectEntry(map[int]string{
		0: "Equal",
		1: "Less Than",
		2: "Greater Than",
	})

	title := "Adding New Watcher"

	pdt := JT.UsePanelMaps().GetDataByID(uuid)
	wk := pdt.UseWatcherKey()

	var parent JW.DialogForm
	var isDisabled = true

	var sent int = wk.GetSent()
	var limit int = wk.GetLimit()

	if !wk.IsEmpty() {

		if sent != -9999 {
			isDisabled = false

		}

		title = "Editing Watcher"
	}

	oe.SetDefaultValue(wk.GetOperator())
	re.SetDefaultValue(strconv.FormatFloat(wk.GetRate(), 'f', -1, 64))
	le.SetDefaultValue(strconv.Itoa(limit))
	de.SetDefaultValue(strconv.Itoa(wk.GetDuration()))

	var bannerBox = container.NewVBox()
	if isDisabled {
		bannerBox.Add(JW.NewBanner(
			"This watcher is disabled. Click the enable button to activate",
			JW.BannerDanger))

		oe.Disable()
		re.Disable()
		le.Disable()
		de.Disable()

	} else if sent > limit {
		bannerBox.Add(JW.NewBanner(
			"This watcher reached its sent limit. Hit Save to reset the limit.",
			JW.BannerWarning))
	} else {
		bannerBox.RemoveAll()
	}

	oe.Validator = validateOps
	re.Validator = validateFloat
	le.Validator = validateInt
	de.Validator = validateInt

	spacer := canvas.NewRectangle(nil)
	spacer.SetMinSize(fyne.NewSize(10, 10))

	duration := container.NewBorder(nil, spacer, nil, widget.NewLabel("Minutes"), de)

	// TODO: Create custom layout to normalize Selector!
	ops := container.NewBorder(oe, spacer, nil, nil)

	fi := []*widget.FormItem{
		widget.NewFormItem("Operator", ops),
		widget.NewFormItem("Rate", re),
		widget.NewFormItem("Limit", le),
		widget.NewFormItem("Duration", duration),
	}

	var label string
	var state string
	if isDisabled {
		label = "Enable"
		state = JW.ActionStateActive
	} else {
		label = "Disable"
		state = JW.ActionStateError
	}

	resetBtn := JW.NewActionButton(
		"disable_watcher",
		label,
		theme.MediaStopIcon(),
		"Disable Watcher",
		state,
		func(btn JW.ActionButton) {
			if !isDisabled {

				oe.Disable()
				re.Disable()
				le.Disable()
				de.Disable()

				btn.SetText("Enable")
				btn.Active()

				isDisabled = true
				sent = -9999

				bannerBox.RemoveAll()
				bannerBox.Add(JW.NewBanner(
					"This watcher is disabled. Click the enable button to activate",
					JW.BannerDanger,
				))

			} else {

				oe.Enable()
				re.Enable()
				le.Enable()
				de.Enable()

				btn.SetText("Disable")
				btn.Error()

				isDisabled = false
				sent = 0

				bannerBox.RemoveAll()
			}

			parent.Refresh()
		},
		nil,
	)

	parent = JW.NewDialogForm(title, fi, []*fyne.Container{bannerBox}, nil, nil, resetBtn,
		func() bool {
			defer func() { allowValidation = false }()
			allowValidation = true

			if isDisabled {
				allowValidation = false
			}

			hasError := false
			if oe.Validate() != nil ||
				re.Validate() != nil ||
				le.Validate() != nil ||
				de.Validate() != nil {
				hasError = true
			}

			if hasError {
				return false
			}

			pdt := JT.UsePanelMaps().GetDataByID(uuid)
			wk := JT.NewWatcherKey()
			pdt.SetWatcherKey(wk.GenerateKeyFromArgs(
				sent,
				oe.GetInt(),
				re.GetFloat(),
				le.GetInt(),
				de.GetInt(),
				0,
			))

			JC.Logln("Watcher check saved:", pdt.UseWatcherKey().GetRawValue())

			// Debug
			// pdt.ProcessWatcher()

			onSave(pdt)

			return true
		},
		onRender,
		func(layer *fyne.Container) {
			if onDestroy != nil {
				onDestroy(layer)
			}
		},
		JC.Window)

	return parent
}
