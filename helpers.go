package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func UpdateDisplay() bool {

	if !JA.AppStatusManager.Refresh().ValidConfig() {
		JC.Logln("Invalid configuration, cannot refresh display")
		JC.Notify("Unable to refresh display: invalid configuration.")

		return false
	}

	list := JT.BP.Get()
	for _, pot := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetData(pot.ID)
		pk := pkt.Get()
		pkt.Update(pk)

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	JC.Logln("Display Refreshed")

	return true
}

func UpdateRates() bool {

	if !JA.AppStatusManager.Refresh().ValidConfig() {
		JC.Logln("Invalid configuration, cannot refresh rates")
		JC.Notify("Unable to refresh rates: invalid configuration.")

		return false
	}

	// Clear cached rates
	JT.ExchangeCache.Reset()

	ex := JT.ExchangeResults{}
	jb := make(map[string]string)
	list := JT.BP.Get()

	// Prune data first, remove duplicate calls, merge into single call wheneveer possible
	for _, pot := range list {
		pk := JT.BP.GetData(pot.ID)
		pkt := pk.UsePanelKey()
		sid := pkt.GetSourceCoinString()
		tid := pkt.GetTargetCoinString()

		_, exists := jb[sid]
		if !exists {
			jb[sid] = sid + "|" + tid
		} else {
			jb[sid] += "," + tid

		}
	}

	if len(jb) == 0 {
		JC.Notify("No valid panels found. Exchange rates were not updated.")
		return false
	}

	JC.Notify("Fetching the latest exchange rates...")

	JA.AppStatusManager.StartFetchingRates()

	// Fetching with delay
	for _, rk := range jb {
		ex.GetRate(rk)

		RequestDisplayUpdate(false)

		// Give pause to prevent too many connection open at once
		time.Sleep(200 * time.Millisecond)
	}

	JC.Notify("Exchange rates updated successfully")

	JC.Logf("Exchange Rate updated: %v/%v", len(jb), len(list))

	JA.AppStatusManager.EndFetchingRates()
	JA.AppStatusManager.Refresh()

	return true
}

func ValidateCache() bool {

	list := JT.BP.Get()
	for _, pot := range list {

		// Always get linked data! do not use the copied
		pkt := JT.BP.GetData(pot.ID)
		pks := pkt.UsePanelKey()
		ck := JT.ExchangeCache.CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

		if !JT.ExchangeCache.Has(ck) {
			return false
		}
	}

	return true
}

func RemovePanel(uuid string) {

	for _, obj := range JC.Grid.Objects {
		if panel, ok := obj.(*JP.PanelDisplay); ok {
			if panel.GetTag() == uuid {

				JC.Logf("Removing panel %s", uuid)

				JC.Grid.Remove(obj)

				fyne.Do(JC.Grid.Refresh)

				if JT.BP.Remove(uuid) {
					if JT.SavePanels() {
						JC.Notify("Panel removed successfully.")
					}
				}

			}
		}
	}

	JA.AppStatusManager.Refresh()
}

func SavePanelForm() {

	JC.Notify("Saving panel settings...")

	JC.Grid.Refresh()
	RequestDisplayUpdate(true)

	go func() {
		if JT.SavePanels() {

			// Only fetch new rates if no cache exists!
			if !ValidateCache() {
				RequestRateUpdate(false)
			}

			JC.Notify("Panel settings saved.")

		} else {
			JC.Notify("Failed to save panel settings.")
		}
	}()
}

func OpenNewPanelForm() {
	fyne.Do(func() {
		d := JP.NewPanelForm(
			"new",
			"",
			SavePanelForm,
			func(npdt *JT.PanelDataType) {

				JC.Grid.Add(CreatePanel(npdt))
				JC.Grid.Refresh()
				JA.AppStatusManager.Refresh()

				JC.Notify("New panel created.")
			},
		)

		d.Show()
		d.Resize(fyne.NewSize(400, 300))
	})
}

func OpenPanelEditForm(pk string, uuid string) {
	fyne.Do(func() {
		d := JP.NewPanelForm(pk, uuid, SavePanelForm, nil)

		d.Show()
		d.Resize(fyne.NewSize(400, 300))
	})
}

func OpenSettingForm() {
	fyne.Do(func() {
		d := JA.NewSettingsForm(func() {
			JC.Notify("Saving configuration...")

			if JT.Config.SaveFile() != nil {
				JC.Notify("Configuration saved successfully.")
			} else {
				JC.Notify("Failed to save configuration.")
			}

			JA.AppStatusManager.Refresh()
		})

		d.Show()
		d.Resize(fyne.NewSize(400, 300))
	})
}

func ToggleDraggable() {

	if JA.AppStatusManager.IsDraggable() {
		JA.AppStatusManager.DisallowDragging()
	} else {
		JA.AppStatusManager.AllowDragging()
	}

	fyne.Do(func() {
		JC.Grid.Refresh()
	})

	JA.AppStatusManager.Refresh()
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanel)
}

func ResetCryptosMap() {
	if !JA.AppStatusManager.Refresh().ValidConfig() {
		JC.Logln("Invalid configuration, cannot reset cryptos map")
		JC.Notify("Invalid configuration. Unable to reset cryptos map.")
		return
	}

	if JA.AppStatusManager.IsFetchingCryptos() {
		return
	}

	JA.AppStatusManager.StartFetchingCryptos()

	Cryptos := JT.CryptosType{}
	JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
	JT.BP.Maps.ClearMapCache()

	if JA.AppStatusManager.Refresh().ValidCryptos() {
		JC.Notify("Cryptos map has been regenerated")
	}

	if JT.BP.RefreshData() {
		fyne.Do(func() {
			JC.Grid.Refresh()
		})

		RequestRateUpdate(false)

	}

	JA.AppStatusManager.EndFetchingCryptos()
	JA.AppStatusManager.Refresh()
}

func StartWorkers() {
	go func() {
		var displayLock sync.Mutex

		for range JC.UpdateDisplayChan {
			displayLock.Lock()

			if UpdateDisplay() {
				JC.UpdateDisplayTimestamp = time.Now()
			}

			displayLock.Unlock()
		}
	}()

	go func() {
		var displayLock sync.Mutex

		for range JC.UpdateRatesChan {
			displayLock.Lock()

			UpdateRates()

			displayLock.Unlock()
		}
	}()
}

func StartUpdateRatesWorker() {
	go func() {
		for {
			RequestRateUpdate(false)
			time.Sleep(time.Duration(JT.Config.Delay) * time.Second)
		}
	}()
}

func RequestDisplayUpdate(force bool) {
	if JT.ExchangeCache.Timestamp.After(JC.UpdateDisplayTimestamp) && JT.ExchangeCache.HasData() || force {
		JC.UpdateDisplayChan <- struct{}{}
	}
}

func RequestRateUpdate(debounce bool) {
	if !JA.AppStatusManager.ValidPanels() {
		return
	}

	if debounce {
		JC.MainDebouncer.Call("update_rates", 1000*time.Millisecond, func() {
			JC.UpdateRatesChan <- struct{}{}
		})
	} else {
		JC.UpdateRatesChan <- struct{}{}
	}
}

func RegisterActions() {
	// Refresh ticker data
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_cryptos", "", theme.ViewRestoreIcon(), "Refresh ticker data",
		func(btn *JW.HoverCursorIconButton) {
			go ResetCryptosMap()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.ValidConfig() {
				btn.Disable()
			} else if !JA.AppStatusManager.ValidCryptos() {
				if JA.AppStatusManager.IsFetchingCryptos() {
					btn.ChangeState("in_progress")
				} else {
					btn.ChangeState("error")
				}
			} else {

				if JA.AppStatusManager.IsFetchingCryptos() {
					btn.ChangeState("in_progress")
				} else {
					btn.ChangeState("reset")
				}
			}
		}))

	// Refresh exchange rates
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_rates", "", theme.ViewRefreshIcon(), "Update rates from exchange",
		func(btn *JW.HoverCursorIconButton) {
			go RequestRateUpdate(true)
		},
		func(btn *JW.HoverCursorIconButton) {
			if JA.AppStatusManager.ValidConfig() && JA.AppStatusManager.ValidCryptos() && JA.AppStatusManager.ValidPanels() {
				if JA.AppStatusManager.IsFetchingRates() {
					btn.ChangeState("in_progress")
				} else {
					btn.Enable()
				}
			} else {
				btn.Disable()
			}
		}))

	// Open settings
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("open_settings", "", theme.SettingsIcon(), "Open settings",
		func(btn *JW.HoverCursorIconButton) {
			go OpenSettingForm()
		},
		func(btn *JW.HoverCursorIconButton) {
			if JA.AppStatusManager.ValidConfig() {
				btn.ChangeState("reset")
			} else {
				btn.ChangeState("error")
			}
		}))

	// Panel drag toggle
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("toggle_drag", "", theme.ContentPasteIcon(), "Enable Reordering",
		func(btn *JW.HoverCursorIconButton) {
			go ToggleDraggable()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.ValidPanels() || JT.BP.TotalData() <= 1 {
				JA.AppStatusManager.DisallowDragging()
				btn.Disable()
			} else if JA.AppStatusManager.IsDraggable() {
				btn.ChangeState("active")
			} else {
				btn.ChangeState("reset")
			}
		}))

	// Add new panel
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("add_panel", "", theme.ContentAddIcon(), "Add new panel",
		func(btn *JW.HoverCursorIconButton) {
			go OpenNewPanelForm()
		},
		func(btn *JW.HoverCursorIconButton) {
			if JA.AppStatusManager.ValidCryptos() {
				btn.Enable()
			} else {
				btn.Disable()
			}
		}))
}
