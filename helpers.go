package main

import (
	"context"
	"fmt"
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

	list := JT.BP.Get()
	for _, pot := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetData(pot.ID)
		pk := pkt.Get()
		pkt.Update(pk)

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	return true
}

func UpdateRates() bool {

	if JA.AppStatusManager.IsFetchingRates() {
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
		return false
	}

	var payloads []any
	for _, rk := range jb {
		payloads = append(payloads, rk)
	}

	var hasError int = 0
	successCount := 0

	JC.Notify("Fetching the latest exchange rates...")
	JA.AppStatusManager.StartFetchingRates()

	JC.FetcherManager.GroupPayloadCall("rates", payloads, func(results []JC.FetchResult) {
		defer JA.AppStatusManager.EndFetchingRates()

		for _, result := range results {
			JT.ExchangeCache.SoftReset()

			ns := DetectHTTPResponse(result.Code)
			if hasError == 0 || hasError < ns {
				hasError = ns
			}
			if ns == 0 {
				successCount++
			}

			JC.WorkerManager.Call("update_display", JC.CallBypassImmediate)
		}

		JC.Logln("Fetching has error:", hasError)

		ProcessUpdatePanelComplete(hasError)

		JC.Logf("Exchange Rate updated: %v/%v", successCount, len(payloads))
	})

	return true
}

func UpdateTickers() bool {

	if JA.AppStatusManager.IsFetchingTickers() {
		return false
	}

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
		return false
	}

	var hasError int = 0

	JC.Notify("Fetching the latest ticker data...")
	JA.AppStatusManager.StartFetchingTickers()
	JT.TickerCache.SoftReset()

	JC.FetcherManager.GroupCall(keys, payloads, func(results map[string]JC.FetchResult) {
		defer JA.AppStatusManager.EndFetchingTickers()

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

		ProcessUpdateTickerComplete(hasError)
	})

	return true
}

func ProcessTickerStatus(status int, tkt *JT.TickerDataType) {
	switch status {
	case 0:
		tkt.Update()
	case 1:
		if !JT.TickerCache.HasData() && tkt.Status != JC.STATE_LOADED && !tkt.HasData() {
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

func ProcessUpdatePanelComplete(status int) {
	switch status {
	case 0:
		JC.Notify("Exchange rates updated successfully")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(true)
		JA.AppSnapshotManager.SavePanels()
		JA.AppSnapshotManager.SaveExchangeData()
	case 1:
		JC.Notify("Please check your network connection.")
		JA.AppStatusManager.SetNetworkStatus(false)

		JC.MainDebouncer.Call("process_rates_complete", 100*time.Millisecond, func() {
			JT.BP.ChangeStatus(JC.STATE_ERROR, func(pdt *JT.PanelDataType) bool {
				return pdt.UsePanelKey().GetValueFloat() < 0
			})

			fyne.Do(func() {
				JP.Grid.UpdatePanelsContent(func(pdt *JT.PanelDataType) bool {
					return true
				})
			})
		})
	case 2:
		JC.Notify("Please check your settings.")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(false)

		JC.MainDebouncer.Call("process_rates_complete", 100*time.Millisecond, func() {
			JT.BP.ChangeStatus(JC.STATE_ERROR, func(pdt *JT.PanelDataType) bool {
				return pdt.UsePanelKey().GetValueFloat() < 0
			})

			fyne.Do(func() {
				JP.Grid.UpdatePanelsContent(func(pdt *JT.PanelDataType) bool {
					return true
				})
			})
		})
	case 3:
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(true)
	}
}

func ProcessUpdateTickerComplete(status int) {
	switch status {
	case 0:
		JC.Notify("Ticker rates updated successfully")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(true)
		JA.AppSnapshotManager.SaveTickers()
		JA.AppSnapshotManager.SaveTickerData()
	case 1:
		JC.Notify("Please check your network connection.")
		JA.AppStatusManager.SetNetworkStatus(false)
		JA.AppStatusManager.SetConfigStatus(true)

		JC.MainDebouncer.Call("process_tickers_complete", 30*time.Millisecond, func() {
			JT.BT.ChangeStatus(JC.STATE_ERROR, func(pdt *JT.TickerDataType) bool {
				return !pdt.HasData()
			})

			fyne.Do(func() {
				JX.Grid.UpdateTickersContent(func(pdt *JT.TickerDataType) bool {
					return true
				})
			})
		})
	case 2:
		JC.Notify("Please check your settings.")
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(false)

		JC.MainDebouncer.Call("process_tickers_complete", 30*time.Millisecond, func() {
			JT.BT.ChangeStatus(JC.STATE_ERROR, func(pdt *JT.TickerDataType) bool {
				return !pdt.HasData()
			})

			fyne.Do(func() {
				JX.Grid.UpdateTickersContent(func(pdt *JT.TickerDataType) bool {
					return true
				})
			})
		})
	case 3:
		JA.AppStatusManager.SetNetworkStatus(true)
		JA.AppStatusManager.SetConfigStatus(true)
	}
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

				if JT.BP.Remove(uuid) {
					fyne.Do(func() {
						JP.Grid.ForceRefresh()

						// Give time for grid to relayout first!
						JC.MainDebouncer.Call("removing_panel", 50*time.Millisecond, func() {
							JA.AppLayoutManager.RefreshLayout()
						})
					})

					if JT.SavePanels() {
						JC.Notify("Panel removed successfully.")
					}
				}

			}
		}
	}

	JA.AppStatusManager.DetectData()
}

func SavePanelForm(pdt *JT.PanelDataType) {

	JC.Notify("Saving panel settings...")

	JP.Grid.ForceRefresh()
	JC.WorkerManager.Call("update_display", JC.CallBypassImmediate)

	go func() {
		if JT.SavePanels() {

			// Only fetch new rates if no cache exists!
			if !ValidateCache() {

				// Force refresh without fail!
				pkt := pdt.UsePanelKey()
				sid := pkt.GetSourceCoinString()
				tid := pkt.GetTargetCoinString()

				payloads := []any{sid + "|" + tid}

				JC.FetcherManager.GroupPayloadCall("rates", payloads, func(results []JC.FetchResult) {
					for _, result := range results {
						JT.ExchangeCache.SoftReset()

						status := DetectHTTPResponse(result.Code)

						switch status {
						case 0:
							opk := pdt.Get()
							if opk != "" {
								pdt.Update(opk)
							}
						case 1, 2, 3:
							pdt.Status = JC.STATE_ERROR
						}

						ProcessUpdatePanelComplete(status)
					}
				})
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
			func(npdt *JT.PanelDataType) {
				SavePanelForm(npdt)
			},
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
		d := JP.NewPanelForm(pk, uuid, func(npdt *JT.PanelDataType) {
			SavePanelForm(npdt)
		}, nil)

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
						JX.Grid = JX.NewTickerGrid()
					}

					JA.AppStatusManager.SetConfigStatus(true)

					JT.TickerCache.SoftReset()
					JC.WorkerManager.Call("update_tickers", JC.CallQueued)

					JT.ExchangeCache.SoftReset()
					JC.WorkerManager.Call("update_rates", JC.CallQueued)
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
		JP.Grid.ForceRefresh()
		if JP.ActiveAction != nil {
			JP.ActiveAction.HideTarget()
		}
	})
}

func ScheduledNotificationReset() {
	JC.MainDebouncer.Call("notification_clear", 6000*time.Millisecond, func() {
		JC.Logln("calling notification clear 1")
		nc, ok := JC.NotificationContainer.(*JW.NotificationDisplayWidget)
		if !ok {
			JC.Logln("calling notification clear 3")
			return
		}

		// Break loop once notification is empty
		if nc.GetText() == "" {
			return
		}

		// Ensure message shown for at least 6 seconds
		last := JC.WorkerManager.GetLastUpdate("notification")
		if time.Since(last) > 6*time.Second {
			JC.Logln("Clearing notification display due to inactivity")
			fyne.Do(func() {
				nc.ClearText()
			})
		} else {
			ScheduledNotificationReset()
		}
	})
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanel)
}

func RegisterActions() {
	// Refresh ticker data
	JA.AppActionManager.AddButton(JW.NewHoverCursorIconButton("refresh_cryptos", "", theme.ViewRestoreIcon(), "Refresh cryptos data",
		func(btn *JW.HoverCursorIconButton) {
			JC.FetcherManager.Call("cryptos_map", nil)
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

			// Open the network status temporarily
			JA.AppStatusManager.SetNetworkStatus(true)

			// Force update
			JT.ExchangeCache.SoftReset()
			JC.WorkerManager.Call("update_rates", JC.CallDebounced)

			// Force update
			JT.TickerCache.SoftReset()
			JC.WorkerManager.Call("update_tickers", JC.CallDebounced)

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

func RegisterWorkers() {

	tickerDelay := max(JT.Config.Delay*1000, 300000)
	ratesDelay := max(JT.Config.Delay*1000, 30000)

	JC.WorkerManager.RegisterSleeper("update_display", func() {
		if UpdateDisplay() {
			JC.UpdateDisplayTimestamp = time.Now()
		}
	}, 200, func() bool {

		if !JA.AppStatusManager.ValidConfig() {
			JC.Logln("Unable to refresh display: invalid configuration")
			return false
		}

		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to refresh display: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to refresh display: app is paused")
			return false
		}

		if !JT.ExchangeCache.HasData() {
			JC.Notify("Unable to refresh display: no cached data")
			return false
		}

		if !JT.ExchangeCache.Timestamp.After(JC.UpdateDisplayTimestamp) {
			JC.Notify("Unable to refresh display: Data is newer than display timestamp")
			return false
		}

		if !JA.AppStatusManager.ValidPanels() {
			JC.Logln("Unable to refresh display: No valid panels configured")
			return false
		}

		return true
	})

	JC.WorkerManager.Register("update_rates", func() {
		UpdateRates()
	}, ratesDelay, 1000, func() bool {

		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to refresh rates: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to refresh rates: app is paused")
			return false
		}

		if !JA.AppStatusManager.ValidConfig() {
			JC.Notify("Unable to refresh rates: invalid configuration")
			return false
		}

		if !JT.ExchangeCache.ShouldRefresh() {
			JC.Logln("Unable to refresh rates: not cleared should refresh yet")
			return false
		}

		if !JA.AppStatusManager.ValidPanels() {
			JC.Logln("Unable to refresh rates: No valid panels configured")
			return false
		}

		// Dont block this, when apps running and connection got dropped,
		// It wont recover due to it is locked!
		// if !JA.AppStatusManager.IsGoodNetworkStatus() {
		// 	JC.Logln("Unable to refresh rates: Network status is bad")
		// 	return false
		// }

		return true
	})

	JC.WorkerManager.Register("update_tickers", func() {
		UpdateTickers()
	}, tickerDelay, 5000, func() bool {

		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to refresh tickers: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to refresh tickers: app is paused")
			return false
		}

		if !JT.Config.IsValidTickers() {
			JC.Logln("Unable to refresh tickers: Invalid ticker configuration")
			return false
		}

		if !JT.TickerCache.ShouldRefresh() {
			JC.Logln("Unable to refresh tickers: Ticker cache shouldn't be refreshed yet")
			return false
		}

		// Dont block this, when apps running and connection got dropped,
		// It wont recover due to it is locked!
		// if !JA.AppStatusManager.IsGoodNetworkStatus() {
		// 	JC.Logln("Unable to refresh tickers: Network status is bad")
		// 	return false
		// }

		return true
	})

	JC.WorkerManager.RegisterBuffered("notification", func(messages []string) bool {
		if len(messages) == 0 {
			return false
		}
		latest := messages[len(messages)-1]

		nc, ok := JC.NotificationContainer.(*JW.NotificationDisplayWidget)
		if !ok {
			return false
		}

		fyne.Do(func() {
			nc.UpdateText(latest)
		})

		ScheduledNotificationReset()

		return true

	}, 1000, 100, 1000, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to do notification: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to do notification: app is paused")
			return false
		}

		return true
	})
}

func RegisterFetchers() {

	var delay int64 = 100

	JC.FetcherManager.Register("cryptos_map", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			c := JT.CryptosType{}
			code := c.GetCryptos()
			return JC.FetchResult{
				Code: code,
				Data: c,
			}, nil
		},
	}, 0, func(result JC.FetchResult) {

		defer JA.AppStatusManager.EndFetchingCryptos()

		code := DetectHTTPResponse(result.Code)

		if code != 0 {
			switch code {
			case 1:
				JC.Notify("Please check your network connection.")
				JA.AppStatusManager.SetNetworkStatus(false)
				JA.AppStatusManager.SetConfigStatus(true)
			case 2:
				JC.Notify("Please check your settings.")
				JA.AppStatusManager.SetConfigStatus(false)
				JA.AppStatusManager.SetNetworkStatus(true)
			case 3:
				JA.AppStatusManager.SetNetworkStatus(true)
				JA.AppStatusManager.SetConfigStatus(true)
			}

			JA.AppStatusManager.SetCryptoStatus(false)

			return
		}

		cryptos, ok := result.Data.(JT.CryptosType)
		if !ok {
			JC.Notify("Invalid crypto data format")
			JA.AppStatusManager.SetCryptoStatus(false)
			return
		}

		CM := cryptos.ConvertToMap()
		if CM == nil {
			JC.Notify("Failed to convert crypto data to map")
			JA.AppStatusManager.SetCryptoStatus(false)
			return
		}

		JT.BP.SetMaps(CM)
		JT.BP.Maps.ClearMapCache()
		JA.AppStatusManager.DetectData()

		if JA.AppStatusManager.ValidCryptos() {
			JC.Notify("Crypto map regenerated successfully")

			JA.AppSnapshotManager.SaveCryptos()

			if JT.BP.RefreshData() {
				fyne.Do(func() {
					JP.Grid.ForceRefresh()
				})

				JT.ExchangeCache.SoftReset()
				JC.WorkerManager.Call("update_rates", JC.CallQueued)

				JT.TickerCache.SoftReset()
				JC.WorkerManager.Call("update_tickers", JC.CallQueued)

				JA.AppStatusManager.SetCryptoStatus(true)
				JA.AppStatusManager.SetConfigStatus(true)
				JA.AppStatusManager.SetNetworkStatus(true)
			}
		}

	}, func() bool {

		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch cryptos: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch cryptos: app is paused")
			return false
		}

		if !JA.AppStatusManager.ValidConfig() {
			JC.Notify("Invalid configuration. Unable to reset cryptos map.")
			JC.Logln("Unable to do fetch cryptos: Invalid config")
			return false
		}

		if JA.AppStatusManager.IsFetchingCryptos() {
			JC.Logln("Unable to do fetch cryptos: Another fetcher is running")
			return false
		}

		JA.AppStatusManager.StartFetchingCryptos()

		return true
	})

	JC.FetcherManager.Register("cmc100", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.CMC100Fetcher{}
			code := ft.GetRate()
			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch CMC100: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch CMC100: app is paused")
			return false
		}

		if !JT.Config.CanDoCMC100() {
			JC.Logln("Unable to fetch CMC100: Invalid config")
			return false
		}

		return true
	})

	JC.FetcherManager.Register("feargreed", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.FearGreedFetcher{}
			code := ft.GetRate()
			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch fear greed: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch fear greed app is paused")
			return false
		}

		if !JT.Config.CanDoFearGreed() {
			JC.Logln("Unable to fetch fear greed: Invalid config")
			return false
		}

		return true
	})

	JC.FetcherManager.Register("market_cap", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.MarketCapFetcher{}
			code := ft.GetRate()
			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch marketcap: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch marketcap: app is paused")
			return false
		}

		if !JT.Config.CanDoMarketCap() {
			JC.Logln("Unable to fetch marketcap: Invalid config")
			return false
		}

		return true
	})

	JC.FetcherManager.Register("altcoin_index", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.AltSeasonFetcher{}
			code := ft.GetRate()
			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch altcoin index: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch altcoin index: app is paused")
			return false
		}

		if !JT.Config.CanDoAltSeason() {
			JC.Logln("Unable to fetch altcoin index: Invalid config")
			return false
		}

		return true
	})

	JC.FetcherManager.Register("rates", &JC.DynamicPayloadFetcher{
		Handler: func(ctx context.Context, payload any) (JC.FetchResult, error) {
			rk, ok := payload.(string)
			if !ok {
				return JC.FetchResult{Code: JC.NETWORKING_BAD_PAYLOAD}, fmt.Errorf("invalid rk")
			}
			ex := &JT.ExchangeResults{}
			code := ex.GetRate(rk)
			return JC.FetchResult{Code: code, Data: ex}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.AppStatusManager.IsReady() {
			JC.Logln("Unable to fetch rates: app is not ready yet")
			return false
		}

		if JA.AppStatusManager.IsPaused() {
			JC.Logln("Unable to fetch rates: app is paused")
			return false
		}

		if !JA.AppStatusManager.ValidPanels() {
			JC.Logln("Unable to rates: no configured panels")
			return false
		}

		return true
	})
}
