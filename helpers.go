package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JM "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
)

func UpdateDisplay() bool {

	if !JT.Config.IsValid() {
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

	if !JT.Config.IsValid() {
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

	// Fetching with delay
	for _, rk := range jb {
		ex.GetRate(rk)

		RequestDisplayUpdate(false)

		// Give pause to prevent too many connection open at once
		time.Sleep(200 * time.Millisecond)
	}

	JC.Notify("Exchange rates updated successfully")

	JC.Logf("Exchange Rate updated: %v/%v", len(jb), len(list))

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
				fyne.Do(JM.AppLayoutManager.Refresh)

				if JT.BP.Remove(uuid) {
					if JT.SavePanels() {
						JC.Notify("Panel removed successfully.")
					}
				}
			}
		}
	}
}

func SavePanelForm() {

	JC.Notify("Saving panel settings...")

	fyne.DoAndWait(func() {
		JC.Grid.Refresh()
		RequestDisplayUpdate(true)
	})

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
				r := len(JC.Grid.Objects) == 0

				fyne.DoAndWait(func() {
					JC.Grid.Add(CreatePanel(npdt))
				})

				if r {
					fyne.DoAndWait(JA.AppLayoutManager.Refresh)
				}

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
		})

		d.Show()
		d.Resize(fyne.NewSize(400, 300))
	})
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanel)
}

func ResetCryptosMap() {
	if !JT.Config.IsValid() {
		JC.Logln("Invalid configuration, cannot reset cryptos map")
		JC.Notify("Invalid configuration. Unable to reset cryptos map.")
		return
	}

	JC.MainDebouncer.Call("update_cryptos", 1000*time.Millisecond, func() {
		Cryptos := JT.CryptosType{}
		JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
		JT.BP.Maps.ClearMapCache()

		JC.Notify("Cryptos map has been regenerated")

		if JT.BP.RefreshData() {
			fyne.DoAndWait(func() {
				JC.Grid.Refresh()
			})

			RequestRateUpdate(false)
		}
	})
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
	if debounce {
		JC.MainDebouncer.Call("update_rates", 1000*time.Millisecond, func() {
			JC.UpdateRatesChan <- struct{}{}
		})
	} else {
		JC.UpdateRatesChan <- struct{}{}
	}
}
