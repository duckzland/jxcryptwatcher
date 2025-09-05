package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func UpdateDisplay() bool {

	if !JA.AppStatusManager.ValidConfig() {
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

	if !JA.AppStatusManager.ValidConfig() {
		JC.Logln("Invalid configuration, cannot refresh rates")
		JC.Notify("Unable to refresh rates: invalid configuration.")

		return false
	}

	if !JT.ExchangeCache.ShouldRefresh() {
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

	if len(jb) == 0 && JA.AppStatusManager.IsReady() {
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

	return true
}

func UpdateTickers() bool {

	if !JT.Config.IsValidTickers() {
		JC.Logln("Invalid ticker configuration, cannot refresh tickers")
		JC.Notify("Unable to refresh tickers: invalid configuration.")

		return false
	}

	// Updating Cache
	cmc100HasError := false
	feargreedHasError := false
	metricsHasError := false
	listingsHasError := false

	if JT.TickerCache.ShouldRefresh() {
		JC.Notify("Fetching the latest ticker data...")
		JA.AppStatusManager.StartFetchingRates()

		// Clear cached rates
		JT.TickerCache.Reset()

		if JA.AppStatusManager.IsValidProKey() {
			cc := JT.TickerCMC100Fetcher{}
			if !DetectProKeyValidityViaHTTPResponse(cc.GetRate()) {
				cmc100HasError = true
			}
		}

		if JA.AppStatusManager.IsValidProKey() {
			fg := JT.TickerFearGreedFetcher{}
			if !DetectProKeyValidityViaHTTPResponse(fg.GetRate()) {
				feargreedHasError = true
			}
		}

		if JA.AppStatusManager.IsValidProKey() {
			mm := JT.TickerMetricsFetcher{}
			if !DetectProKeyValidityViaHTTPResponse(mm.GetRate()) {
				metricsHasError = true
			}
		}

		if JA.AppStatusManager.IsValidProKey() {
			lt := JT.TickerListingsFetcher{}
			if !DetectProKeyValidityViaHTTPResponse(lt.GetRate()) {
				listingsHasError = true
			}
		}

		JA.AppStatusManager.EndFetchingRates()
		if !cmc100HasError && !feargreedHasError && !metricsHasError && !listingsHasError {
			JC.Notify("Ticker data updated successfully")
			JA.AppStatusManager.SetConfigStatus(true)
		} else {
			JC.Notify("Please check your settings and network connection.")
			JA.AppStatusManager.SetConfigStatus(false)
		}
	}

	// Refreshing Display
	list := JT.BT.Get()
	for _, tkt := range list {
		switch tkt.Type {
		case "cmc100":
			if cmc100HasError {
				tkt.Set("-1")
				tkt.Status = -1
			} else {
				tkt.Update()
			}
		case "feargreed":
			if feargreedHasError {
				tkt.Set("-1")
				tkt.Status = -1
			} else {
				tkt.Update()
			}
		case "market_cap":
			if metricsHasError {
				tkt.Set("-1")
				tkt.Status = -1
			} else {
				tkt.Update()
			}
		case "altcoin_index":
			if listingsHasError {
				tkt.Set("-1")
				tkt.Status = -1
			} else {
				tkt.Update()
			}
		}
	}

	return true
}

func DetectProKeyValidityViaHTTPResponse(rs int64) bool {
	switch rs {
	case 401, 403, 429:
		JA.AppStatusManager.SetCryptoKeyStatus(false)
	case 200:
		JA.AppStatusManager.SetCryptoKeyStatus(true)
		return true
	}

	return false
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

	for _, obj := range JP.Grid.Objects {
		if panel, ok := obj.(*JP.PanelDisplay); ok {
			if panel.GetTag() == uuid {

				JC.Logf("Removing panel %s", uuid)

				JP.Grid.Remove(obj)

				fyne.Do(JP.Grid.Refresh)

				if JT.BP.Remove(uuid) {
					if JT.SavePanels() {
						JC.Notify("Panel removed successfully.")
					}
				}

			}
		}
	}

	JA.AppStatusManager.DetectData()
}

func SavePanelForm() {

	JC.Notify("Saving panel settings...")

	JP.Grid.Refresh()
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

				JP.Grid.Add(CreatePanel(npdt))
				JP.Grid.Refresh()
				JA.AppStatusManager.DetectData()

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
				JA.AppStatusManager.DetectData()
				if JT.Config.IsValidTickers() {
					if JT.BT.IsEmpty() {
						JC.Logln("Rebuilding tickers due to empty ticker list")
						JT.TickersInit()
						JC.Tickers = JX.NewTickerGrid()
					}
					JA.AppStatusManager.SetCryptoKeyStatus(true)
					JA.AppStatusManager.SetConfigStatus(true)
					JT.TickerCache.Reset()
					RequestTickersUpdate()
				}
			} else {
				JC.Notify("Failed to save configuration.")
			}
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
		JP.Grid.Refresh()
	})
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanel)
}

func ResetCryptosMap() {
	if !JA.AppStatusManager.ValidConfig() {
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

	JA.AppStatusManager.EndFetchingCryptos()

	if JA.AppStatusManager.ValidCryptos() {
		JC.Notify("Cryptos map has been regenerated")
	}

	if JT.BP.RefreshData() {
		fyne.Do(func() {
			JP.Grid.Refresh()
		})

		RequestRateUpdate(false)

	}
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

	go func() {
		var displayLock sync.Mutex

		for range JC.UpdateTickersChan {
			displayLock.Lock()

			UpdateTickers()

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

func StartUpdateTickersWorker() {
	go func() {
		for {
			RequestTickersUpdate()
			time.Sleep(time.Duration(JT.Config.TickerDelay) * time.Second)
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

func RequestTickersUpdate() {
	if JT.TickerCache.ShouldRefresh() {
		JC.MainDebouncer.Call("update_tickers", 1000*time.Millisecond, func() {
			JC.UpdateTickersChan <- struct{}{}
		})
	}
}

func RegisterActions() {
	// Refresh ticker data
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_cryptos", "", theme.ViewRestoreIcon(), "Refresh ticker data",
		func(btn *JW.HoverCursorIconButton) {
			go ResetCryptosMap()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidConfig() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsFetchingCryptos() {
				btn.Progress()
				return
			}

			if !JA.AppStatusManager.ValidCryptos() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Refresh exchange rates
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_rates", "", theme.ViewRefreshIcon(), "Update rates from exchange",
		func(btn *JW.HoverCursorIconButton) {
			go RequestRateUpdate(true)
			go RequestTickersUpdate()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidConfig() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidCryptos() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidPanels() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsFetchingRates() {
				btn.Progress()
				return
			}

			btn.Enable()
		}))

	// Open settings
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("open_settings", "", theme.SettingsIcon(), "Open settings",
		func(btn *JW.HoverCursorIconButton) {
			go OpenSettingForm()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidConfig() {
				btn.Error()
				return
			}

			if !JA.AppStatusManager.IsValidTickerConfig() {
				btn.Error()
				return
			}

			if !JA.AppStatusManager.IsValidProKey() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Panel drag toggle
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("toggle_drag", "", theme.ContentPasteIcon(), "Enable Reordering",
		func(btn *JW.HoverCursorIconButton) {
			go ToggleDraggable()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.IsReady() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidPanels() {
				JA.AppStatusManager.DisallowDragging()
				btn.Disable()
				return
			}

			if JT.BP.TotalData() < 2 {
				JA.AppStatusManager.DisallowDragging()
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsDraggable() {
				btn.Active()
				return
			}

			btn.Enable()
		}))

	// Add new panel
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("add_panel", "", theme.ContentAddIcon(), "Add new panel",
		func(btn *JW.HoverCursorIconButton) {
			go OpenNewPanelForm()
		},
		func(btn *JW.HoverCursorIconButton) {
			if !JA.AppStatusManager.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatusManager.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatusManager.ValidCryptos() {
				btn.Disable()
				return
			}

			btn.Enable()
		}))
}
