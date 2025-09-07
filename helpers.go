package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"jxwatcher/apps"
	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
	"jxwatcher/widgets"
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
	if !JA.AppStatusManager.ValidConfig() {
		JC.Logln("Invalid configuration, cannot refresh rates")
		JC.Notify("Unable to refresh rates: invalid configuration.")
		return false
	}

	if !JT.ExchangeCache.ShouldRefresh() {
		JC.Logln("Unable to refresh rates: not cleared should refresh yet.")
		return false
	}

	JT.ExchangeCache.SoftReset()

	jb := make(map[string]string)
	list := JT.BP.Get()

	for _, pot := range list {
		pk := JT.BP.GetData(pot.ID)
		pkt := pk.UsePanelKey()
		sid := pkt.GetSourceCoinString()
		tid := pkt.GetTargetCoinString()

		if _, exists := jb[sid]; !exists {
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

	var payloads []any
	for _, rk := range jb {
		payloads = append(payloads, rk)
	}

	var hasError int = 0
	successCount := 0

	JA.AppFetcherManager.GroupPayloadCall("rates", payloads, func(results []JA.AppFetchResult) {
		for _, result := range results {
			JT.ExchangeCache.SoftReset()

			ns := DetectHTTPResponse(result.Code)
			if hasError == 0 || hasError < ns {
				hasError = ns
			}
			if ns == 0 {
				successCount++
			}
			JA.AppWorkerManager.Call("update_display", JA.CallBypassImmediate)
		}

		JP.Grid.UpdatePanelsContent()
		JC.Logln("Fetching has error:", hasError)

		JA.AppWorkerManager.Call("update_display", JA.CallBypassImmediate)

		switch hasError {
		case 0:
			JC.Notify("Exchange rates updated successfully")
			JA.AppStatusManager.SetNetworkStatus(true)
		case 1:
			JC.Notify("Please check your network connection.")
			JA.AppStatusManager.SetNetworkStatus(false)
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
		}

		JC.Logf("Exchange Rate updated: %v/%v", successCount, len(payloads))
		JA.AppStatusManager.EndFetchingRates()
	})

	return true
}

func UpdateTickers() bool {
	if !JT.Config.IsValidTickers() {
		JC.Logln("Invalid ticker configuration, cannot refresh tickers")
		if JA.AppStatusManager.IsReady() {
			JC.Notify("Unable to refresh tickers: invalid configuration.")
		}
		return false
	}

	if !JT.TickerCache.ShouldRefresh() {
		JC.Logln("Unable to refresh tickers: not cleared should refresh yet.")
		return false
	}

	JC.Notify("Fetching the latest ticker data...")
	JA.AppStatusManager.StartFetchingTickers()
	JT.TickerCache.SoftReset()

	// Prepare keys and payloads
	keys := []string{}
	payloads := map[string]any{}

	if JT.Config.CanDoCMC100() {
		keys = append(keys, "cmc100")
		payloads["cmc100"] = nil
	}
	if JT.Config.CanDoFearGreed() {
		keys = append(keys, "feargreed")
		payloads["feargreed"] = nil
	}
	if JT.Config.CanDoMarketCap() {
		keys = append(keys, "market_cap")
		payloads["market_cap"] = nil
	}
	if JT.Config.CanDoAltSeason() {
		keys = append(keys, "altcoin_index")
		payloads["altcoin_index"] = nil
	}

	if len(keys) == 0 {
		JC.Notify("No valid ticker sources available.")
		JA.AppStatusManager.EndFetchingTickers()
		return false
	}

	var hasError int = 0

	JA.AppFetcherManager.GroupCall(keys, payloads, func(results map[string]JA.AppFetchResult) {
		for key, result := range results {
			ns := DetectHTTPResponse(result.Code)
			tktt := JT.BT.GetDataByType(key)

			for _, tkt := range tktt {
				ProcessTickerStatus(ns, tkt)
			}

			if hasError == 0 || hasError < ns {
				hasError = ns
			}
		}

		JA.AppStatusManager.EndFetchingTickers()

		switch hasError {
		case 0:
			JC.Notify("Ticker rates updated successfully")
			JA.AppStatusManager.SetNetworkStatus(true)
		case 1:
			JC.Notify("Please check your network connection.")
			JA.AppStatusManager.SetNetworkStatus(false)
		case 2:
			JC.Notify("Please check your settings.")
			JA.AppStatusManager.SetNetworkStatus(true)
			JA.AppStatusManager.SetConfigStatus(false)
		case 3:
			JA.AppStatusManager.SetNetworkStatus(true)
		}
	})

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

				fyne.Do(JP.Grid.ForceRefresh)

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

	JP.Grid.ForceRefresh()
	JA.AppWorkerManager.Call("update_display", JA.CallBypassImmediate)

	go func() {
		if JT.SavePanels() {

			// Only fetch new rates if no cache exists!
			if !ValidateCache() {
				// Force Refresh
				JT.ExchangeCache.SoftReset()
				JA.AppWorkerManager.Call("update_rates", JA.CallImmediate)
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
				JP.Grid.ForceRefresh()
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

					JT.TickerCache.SoftReset()
					JA.AppWorkerManager.Call("update_tickers", JA.CallImmediate)

					JT.ExchangeCache.SoftReset()
					JA.AppWorkerManager.Call("update_rates", JA.CallImmediate)
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
						JP.Grid.ForceRefresh()
					})

					// Force Refresh
					JT.ExchangeCache.SoftReset()
					JA.AppWorkerManager.Call("update_rates", JA.CallImmediate)

					// Force Refresh
					JT.TickerCache.SoftReset()
					JA.AppWorkerManager.Call("update_tickers", JA.CallImmediate)

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
				JA.AppWorkerManager.Call("update_rates", JA.CallDebounced)

				// Force update
				JT.TickerCache.SoftReset()
				JA.AppWorkerManager.Call("update_tickers", JA.CallDebounced)
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

func SetupWorkers() {
	JA.AppWorkerManager.RegisterSleeper("update_display", func() {
		if UpdateDisplay() {
			JC.UpdateDisplayTimestamp = time.Now()
		}
	}, 200, func() bool {
		return JT.ExchangeCache.Timestamp.After(JC.UpdateDisplayTimestamp) && JT.ExchangeCache.HasData()
	})

	JA.AppWorkerManager.Register("update_rates", func() {
		UpdateRates()
	}, JT.Config.Delay*1000, 1000, func() bool {
		return JA.AppStatusManager.ValidPanels() && JA.AppStatusManager.IsGoodNetworkStatus()
	})

	JA.AppWorkerManager.Register("update_tickers", func() {
		UpdateTickers()
	}, JT.Config.Delay*1000, 1000, func() bool {
		return JT.Config.IsValidTickers() && JA.AppStatusManager.IsGoodNetworkStatus()
	})

	JA.AppWorkerManager.RegisterBuffered("notification", func(messages []string) bool {
		if len(messages) == 0 {
			return false
		}
		latest := messages[len(messages)-1]

		nc, ok := JC.NotificationContainer.(*widgets.NotificationDisplayWidget)
		if !ok {
			return false
		}

		fyne.Do(func() {
			nc.UpdateText(latest)
		})

		return true

	}, 1000, 100, 1000, nil)

	JA.AppWorkerManager.Register("notification_idle_clear", func() {

		nc, ok := JC.NotificationContainer.(*widgets.NotificationDisplayWidget)
		if !ok {
			return
		}

		JC.Logln("Clearing notification display due to inactivity")
		fyne.Do(func() {
			nc.ClearText()
		})

	}, 5000, 1000, func() bool {

		nc, ok := JC.NotificationContainer.(*widgets.NotificationDisplayWidget)
		if !ok {
			return false
		}
		last := JA.AppWorkerManager.GetLastUpdate("notification")
		return time.Since(last) > 6*time.Second && nc.GetText() != ""
	})
}

func SetupFetchers() {

	var delay int64 = 100

	JA.AppFetcherManager.Register("cmc100", &JA.GenericFetcher{
		Handler: func(ctx context.Context) (JA.AppFetchResult, error) {
			ft := JT.CMC100Fetcher{}
			code := ft.GetRate()
			return JA.AppFetchResult{Code: code}, nil
		},
	}, delay, nil)

	JA.AppFetcherManager.Register("feargreed", &JA.GenericFetcher{
		Handler: func(ctx context.Context) (JA.AppFetchResult, error) {
			ft := JT.FearGreedFetcher{}
			code := ft.GetRate()
			return JA.AppFetchResult{Code: code}, nil
		},
	}, delay, nil)

	JA.AppFetcherManager.Register("market_cap", &JA.GenericFetcher{
		Handler: func(ctx context.Context) (JA.AppFetchResult, error) {
			ft := JT.MarketCapFetcher{}
			code := ft.GetRate()
			return JA.AppFetchResult{Code: code}, nil
		},
	}, delay, nil)

	JA.AppFetcherManager.Register("altcoin_index", &JA.GenericFetcher{
		Handler: func(ctx context.Context) (JA.AppFetchResult, error) {
			ft := JT.AltSeasonFetcher{}
			code := ft.GetRate()
			return JA.AppFetchResult{Code: code}, nil
		},
	}, delay, nil)

	JA.AppFetcherManager.Register("rates", &apps.DynamicPayloadFetcher{
		Handler: func(ctx context.Context, payload any) (apps.AppFetchResult, error) {
			rk, ok := payload.(string)
			if !ok {
				return apps.AppFetchResult{Code: JC.NETWORKING_BAD_PAYLOAD}, fmt.Errorf("invalid rk")
			}
			ex := &JT.ExchangeResults{}
			code := ex.GetRate(rk)
			return apps.AppFetchResult{Code: code, Data: ex}, nil
		},
	}, delay, nil)
}
