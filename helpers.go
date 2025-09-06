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

	// JC.UpdateDisplayLock.Lock()
	// defer JC.UpdateDisplayLock.Unlock()

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

		// JC.Logln("Updating: ", pk)

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	// JC.Logln("Display Refreshed")

	return true
}

func UpdateRates() bool {

	// JC.UpdateRatesLock.Lock()
	// defer JC.UpdateRatesLock.Unlock()

	if !JA.AppStatusManager.ValidConfig() {
		JC.Logln("Invalid configuration, cannot refresh rates")
		JC.Notify("Unable to refresh rates: invalid configuration.")

		return false
	}

	if !JT.ExchangeCache.ShouldRefresh() {
		JC.Logln("Unable to refresh rates: not cleared should refresh yet.")
		return false
	}

	// Clear cached rates
	JT.ExchangeCache.SoftReset()

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
		if JA.AppStatusManager.IsReady() {
			JC.Notify("No valid panels found. Exchange rates were not updated.")
		}
		return false
	}

	JC.Notify("Fetching the latest exchange rates...")

	JA.AppStatusManager.StartFetchingRates()

	var hasError int = 0
	// var mu sync.Mutex
	succesCount := 0

	for _, rk := range jb {

		JT.ExchangeCache.SoftReset()
		rs := ex.GetRate(rk)

		// mu.Lock()
		ns := DetectHTTPResponse(rs)
		// JC.Logln("Processing : ", rk, " with ns: ", ns)

		if hasError == 0 || hasError < ns {
			hasError = ns
		}
		if hasError == 0 {
			succesCount++
		}

		RequestDisplayUpdate(true)

		// mu.Unlock()

		// time.Sleep(100 * time.Millisecond)
	}

	JP.Grid.UpdatePanelsContent()

	if hasError != 0 {
		JC.Logln("Error when fetching rates:", hasError)
	}

	switch hasError {
	case 0:
		JC.Notify("Exchange rates updated successfully")
		JA.AppStatusManager.SetNetworkStatus(true)
		// JA.AppStatusManager.SetConfigStatus(true)

	case 1:
		JC.Notify("Please check your network connection.")
		JA.AppStatusManager.SetNetworkStatus(false)
		// JA.AppStatusManager.SetConfigStatus(true)

		if !JT.ExchangeCache.HasData() {
			JP.PanelForceUpdate = true
			JT.BP.ChangeAllStatus(JC.STATE_ERROR)
			JP.Grid.UpdatePanelsContent()
			JP.PanelForceUpdate = false
		}

	case 2:
		JC.Notify("Please check your settings.")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(false)

		if !JT.ExchangeCache.HasData() {
			JP.PanelForceUpdate = true
			JT.BP.ChangeAllStatus(JC.STATE_ERROR)
			JP.Grid.UpdatePanelsContent()
			JP.PanelForceUpdate = false
		}
	case 3:
		JA.AppStatusManager.SetNetworkStatus(true)
		// JA.AppStatusManager.SetConfigStatus(true)
	}

	JC.Logf("Exchange Rate updated: %v/%v", succesCount, len(jb))

	JA.AppStatusManager.EndFetchingRates()

	return true
}

func UpdateTickers() bool {

	// JC.UpdateTickersLock.Lock()
	// defer JC.UpdateTickersLock.Unlock()

	if !JT.Config.IsValidTickers() {
		JC.Logln("Invalid ticker configuration, cannot refresh tickers")

		if JA.AppStatusManager.IsReady() {
			JC.Notify("Unable to refresh tickers: invalid configuration.")
		}

		return false
	}

	var hasError int = 0

	if !JT.TickerCache.ShouldRefresh() {
		JC.Logln("Unable to refresh tickers: not cleared should refresh yet.")
		return false
	}

	// var mu sync.Mutex

	JC.Notify("Fetching the latest ticker data...")
	JA.AppStatusManager.StartFetchingTickers()

	// Clear cached rates
	JT.TickerCache.SoftReset()

	if JT.Config.CanDoCMC100() {
		ft := JT.CMC100Fetcher{}
		JT.TickerCache.SoftReset()
		rs := ft.GetRate()

		tktt := JT.BT.GetDataByType("cmc100")
		ns := DetectHTTPResponse(rs)

		for _, tkt := range tktt {
			ProcessTickerStatus(ns, tkt)
		}

		// mu.Lock()
		if hasError == 0 {
			hasError = ns
		}
		// mu.Unlock()

		// time.Sleep(100 * time.Millisecond)
	}

	if JT.Config.CanDoFearGreed() {
		ft := JT.FearGreedFetcher{}
		JT.TickerCache.SoftReset()
		rs := ft.GetRate()

		tktt := JT.BT.GetDataByType("feargreed")
		ns := DetectHTTPResponse(rs)

		for _, tkt := range tktt {
			ProcessTickerStatus(ns, tkt)
		}

		// mu.Lock()
		if hasError == 0 {
			hasError = ns
		}
		// mu.Unlock()

		// time.Sleep(200 * time.Millisecond)
	}

	if JT.Config.CanDoMarketCap() {

		ft := JT.MarketCapFetcher{}
		JT.TickerCache.SoftReset()
		rs := ft.GetRate()

		tktt := JT.BT.GetDataByType("market_cap")
		ns := DetectHTTPResponse(rs)

		for _, tkt := range tktt {
			ProcessTickerStatus(ns, tkt)
		}

		// mu.Lock()
		if hasError == 0 || hasError < ns {
			hasError = ns
		}
		// mu.Unlock()

		// time.Sleep(200 * time.Millisecond)
	}

	if JT.Config.CanDoAltSeason() {

		ft := JT.AltSeasonFetcher{}
		JT.TickerCache.SoftReset()
		rs := ft.GetRate()

		tktt := JT.BT.GetDataByType("altcoin_index")
		ns := DetectHTTPResponse(rs)

		for _, tkt := range tktt {
			ProcessTickerStatus(ns, tkt)
		}

		// mu.Lock()
		if hasError == 0 {
			hasError = ns
		}
		// mu.Unlock()

		// time.Sleep(200 * time.Millisecond)
	}

	JA.AppStatusManager.EndFetchingTickers()

	switch hasError {
	case 0:
		JC.Notify("Ticker rates updated successfully")
		JA.AppStatusManager.SetNetworkStatus(true)
		// JA.AppStatusManager.SetConfigStatus(true)
	case 1:
		JC.Notify("Please check your network connection.")
		JA.AppStatusManager.SetNetworkStatus(false)
		// JA.AppStatusManager.SetConfigStatus(true)
	case 2:
		JC.Notify("Please check your settings.")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(false)
	case 3:
		JA.AppStatusManager.SetNetworkStatus(true)
		// JA.AppStatusManager.SetConfigStatus(true)
	}

	return true
}

func ProcessTickerStatus(status int, tkt *JT.TickerDataType) {
	switch status {
	case 0:
		tkt.Update()
	case 1:
		if !JT.TickerCache.HasData() {
			tkt.Set("-1")
			tkt.Status = JC.STATE_ERROR
		}
	case 2:
		tkt.Set("-1")
		tkt.Status = JC.STATE_ERROR
	}
}

func DetectHTTPResponse(rs int64) int {

	// JC.Logln("Raw rs value: ", rs)
	switch rs {
	case JC.NETWORKING_SUCCESS:
		return 0

	case JC.NETWORKING_ERROR_CONNECTION:
		return 1

	case JC.NETWORKING_BAD_CONFIG, JC.NETWORKING_URL_ERROR:
		return 2

	case JC.NETWORKING_BAD_DATA_RECEIVED, JC.NETWORKING_DATA_IN_CACHE, JC.NETWORKING_BAD_PAYLOAD, JC.NETWORKING_FAILED_CREATE_FILE:
		return 3

	}

	return 0
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

				JP.ForceLayoutRefresh = true
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

	JP.ForceLayoutRefresh = true
	JP.Grid.Refresh()
	RequestDisplayUpdate(true)

	go func() {
		if JT.SavePanels() {

			// Only fetch new rates if no cache exists!
			if !ValidateCache() {
				// Force Refresh
				JT.ExchangeCache.SoftReset()
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

					JA.AppStatusManager.SetConfigStatus(true)
					JA.AppStatusManager.SetConfigStatus(true)

					JT.TickerCache.SoftReset()
					RequestTickersUpdate()

					JT.ExchangeCache.SoftReset()
					RequestRateUpdate(true)
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

	// Fetch and generate json
	status := DetectHTTPResponse(Cryptos.GetCryptos())

	switch status {
	case 0:
		CM := Cryptos.LoadFile().ConvertToMap()
		if CM != nil {
			JT.BP.SetMaps(CM)
			JT.BP.Maps.ClearMapCache()

			JA.AppStatusManager.DetectData()
			if JA.AppStatusManager.ValidCryptos() {
				JC.Notify("Crypto map regenerated successfully")

				if JT.BP.RefreshData() {
					fyne.Do(func() {
						JP.Grid.Refresh()
					})

					// Force Refresh
					JT.ExchangeCache.SoftReset()
					RequestRateUpdate(false)

					// Force Refresh
					JT.TickerCache.SoftReset()
					RequestTickersUpdate()

					JA.AppStatusManager.SetCryptoStatus(true)

				}
			}
		}
	case 1:
		JC.Notify("Please check your network connection.")
		JA.AppStatusManager.SetCryptoStatus(false)
	case 2, 3:
		JC.Notify("Please check your settings.")
		JA.AppStatusManager.SetCryptoStatus(false)
	}

	JA.AppStatusManager.EndFetchingCryptos()
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

func RequestTickersUpdate() {
	JC.MainDebouncer.Call("update_tickers", 1000*time.Millisecond, func() {
		JC.UpdateTickersChan <- struct{}{}
	})
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

			if !JA.AppStatusManager.IsValidCrypto() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Refresh exchange rates
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_rates", "", theme.ViewRefreshIcon(), "Update rates from exchange",
		func(btn *JW.HoverCursorIconButton) {
			go func() {
				// Force update
				JT.ExchangeCache.SoftReset()
				RequestRateUpdate(true)

				// Force update
				JT.TickerCache.SoftReset()
				RequestTickersUpdate()
			}()
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

			if !JA.AppStatusManager.IsValidConfig() {
				btn.Error()
				return
			}

			if !JA.AppStatusManager.IsGoodNetworkStatus() {
				btn.Error()
				return
			}

			if JA.AppStatusManager.IsFetchingRates() {
				btn.Progress()
				return
			}

			if JA.AppStatusManager.IsFetchingTickers() {
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

			if !JA.AppStatusManager.IsValidConfig() {
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
