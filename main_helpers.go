package main

import (
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func UpdateDisplay() bool {

	list := JT.BP.GetData()
	for _, pot := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetDataByID(pot.GetID())
		pk := pkt.Get()
		pkt.Update(pk)

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	return true
}

func UpdateRates() bool {

	if JA.StatusManager.IsFetchingRates() {
		return false
	}

	jb := make(map[string]string)
	list := JT.BP.GetData()

	for _, pot := range list {
		pk := JT.BP.GetDataByID(pot.GetID())

		if !JT.BP.ValidatePanel(pk.Get()) {
			pk.SetStatus(JC.STATE_BAD_CONFIG)
			continue
		}

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

	JC.FetcherManager.GroupPayloadCall("rates", payloads,
		func(shouldProceed bool) {
			if shouldProceed {
				JA.StatusManager.StartFetchingRates()
				JT.ExchangeCache.SoftReset()
			}
		},
		func(results []JC.FetchResult) {
			defer JA.StatusManager.EndFetchingRates()

			for _, result := range results {

				ns := DetectHTTPResponse(result.Code)
				if hasError == JC.STATUS_SUCCESS || hasError < ns {
					hasError = ns
				}

				if ns == JC.STATUS_SUCCESS {
					successCount++
				}
			}

			if successCount != 0 {
				JC.WorkerManager.Call("update_display", JC.CallBypassImmediate)
			}

			JC.Logln("Fetching has error:", hasError)

			ProcessUpdatePanelComplete(hasError)

			JC.Logf("Exchange Rate updated: %v/%v", successCount, len(payloads))
		})

	return true
}

func UpdateTickers() bool {

	if JA.StatusManager.IsFetchingTickers() {
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

	JC.FetcherManager.GroupCall(keys, payloads,
		func(totalJob int) {
			if totalJob > 0 {
				JA.StatusManager.StartFetchingTickers()
				JT.TickerCache.SoftReset()
			}
		},
		func(results map[string]JC.FetchResult) {
			defer JA.StatusManager.EndFetchingTickers()

			for key, result := range results {
				ns := DetectHTTPResponse(result.Code)
				tktt := JT.BT.GetDataByType(key)

				for _, tkt := range tktt {
					switch ns {
					case JC.STATUS_SUCCESS:
						tkt.Update()
					}
				}

				if hasError == 0 || hasError < ns {
					hasError = ns
				}
			}

			ProcessUpdateTickerComplete(hasError)
		})

	return true
}

func DetectHTTPResponse(rs int64) int {

	switch rs {
	case JC.NETWORKING_SUCCESS:
		return JC.STATUS_SUCCESS

	case JC.NETWORKING_ERROR_CONNECTION:
		return JC.STATUS_NETWORK_ERROR

	case JC.NETWORKING_BAD_CONFIG, JC.NETWORKING_URL_ERROR:
		return JC.STATUS_CONFIG_ERROR

	case JC.NETWORKING_BAD_DATA_RECEIVED, JC.NETWORKING_DATA_IN_CACHE, JC.NETWORKING_BAD_PAYLOAD, JC.NETWORKING_FAILED_CREATE_FILE:
		return JC.STATUS_BAD_DATA_RECEIVED

	}

	return JC.STATUS_SUCCESS
}

func ProcessUpdatePanelComplete(status int) {
	switch status {
	case JC.STATUS_SUCCESS:

		JC.Notify("Exchange rates updated successfully")
		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify("Please check your network connection.")
		JA.StatusManager.SetNetworkStatus(false)

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

	case JC.STATUS_CONFIG_ERROR:

		JC.Notify("Please check your settings.")
		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(false)

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

	case JC.STATUS_BAD_DATA_RECEIVED:

		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(true)
	}
}

func ProcessUpdateTickerComplete(status int) {

	switch status {
	case JC.STATUS_SUCCESS:

		JC.Notify("Ticker rates updated successfully")
		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify("Please check your network connection.")
		JA.StatusManager.SetNetworkStatus(false)
		JA.StatusManager.SetConfigStatus(true)

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

	case JC.STATUS_CONFIG_ERROR:

		JC.Notify("Please check your settings.")

		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(false)

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

	case JC.STATUS_BAD_DATA_RECEIVED:

		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(true)
	}
}

func ProcessFetchingCryptosComplete(status int) {

	switch status {
	case JC.STATUS_SUCCESS:

		JT.CryptosInit()
		JA.StatusManager.DetectData()

		if !JA.StatusManager.ValidCryptos() {
			JC.Notify("Failed to convert crypto data to map")
			JA.StatusManager.SetCryptoStatus(false)

			return
		}

		JC.Notify("Crypto map regenerated successfully")

		if JT.BP.RefreshData() {
			fyne.Do(func() {
				JP.Grid.ForceRefresh()
			})

			JT.ExchangeCache.SoftReset()
			JC.WorkerManager.Call("update_rates", JC.CallQueued)

			JT.TickerCache.SoftReset()
			JC.WorkerManager.Call("update_tickers", JC.CallQueued)

			JA.StatusManager.SetCryptoStatus(true)
			JA.StatusManager.SetConfigStatus(true)
			JA.StatusManager.SetNetworkStatus(true)
		}

	case JC.STATUS_NETWORK_ERROR:
		JC.Notify("Please check your network connection.")
		JA.StatusManager.SetNetworkStatus(false)
		JA.StatusManager.SetConfigStatus(true)

	case JC.STATUS_CONFIG_ERROR:
		JC.Notify("Please check your settings.")
		JA.StatusManager.SetConfigStatus(false)
		JA.StatusManager.SetNetworkStatus(true)

	case JC.STATUS_BAD_DATA_RECEIVED:
		JA.StatusManager.SetNetworkStatus(true)
		JA.StatusManager.SetConfigStatus(true)
	}
}

func ValidateRatesCache() bool {

	list := JT.BP.GetData()
	for _, pot := range list {

		// Always get linked data! do not use the copied
		pkt := JT.BP.GetDataByID(pot.GetID())
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
					JP.Grid.ForceRefresh()

					// Give time for grid to relayout first!
					JC.MainDebouncer.Call("removing_panel", 50*time.Millisecond, func() {
						JA.LayoutManager.RefreshLayout()

						if JT.SavePanels() {
							JC.Notify("Panel removed successfully.")
						}
					})
				}

			}
		}
	}

	JA.StatusManager.DetectData()
}

func SavePanelForm(pdt *JT.PanelDataType) {

	JC.Notify("Saving panel settings...")

	JP.Grid.ForceRefresh()

	if !JT.BP.ValidatePanel(pdt.Get()) {
		pdt.SetStatus(JC.STATE_BAD_CONFIG)
	}

	JC.WorkerManager.Call("update_display", JC.CallBypassImmediate)

	go func() {
		if JT.SavePanels() {

			if pdt.IsStatus(JC.STATE_BAD_CONFIG) {
				return
			}

			// Only fetch new rates if no cache exists!
			if !ValidateRatesCache() {

				// Force refresh without fail!
				pkt := pdt.UsePanelKey()

				sid := pkt.GetSourceCoinString()
				tid := pkt.GetTargetCoinString()

				payloads := []any{sid + "|" + tid}

				JC.FetcherManager.GroupPayloadCall("rates", payloads,
					func(shouldProceed bool) {
					},
					func(results []JC.FetchResult) {
						for _, result := range results {
							JT.ExchangeCache.SoftReset()

							status := DetectHTTPResponse(result.Code)

							switch status {
							case JC.STATUS_SUCCESS:

								opk := pdt.Get()
								if opk != "" {
									pdt.Update(opk)
								}

							case JC.STATUS_NETWORK_ERROR, JC.STATUS_CONFIG_ERROR, JC.STATUS_BAD_DATA_RECEIVED:
								pdt.SetStatus(JC.STATE_ERROR)
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
	if JA.StatusManager.IsOverlayShown() {
		return
	}

	JA.StatusManager.SetOverlayShownStatus(true)

	d := JP.NewPanelForm(
		"new",
		"",
		func(npdt *JT.PanelDataType) {
			SavePanelForm(npdt)
		},
		func(npdt *JT.PanelDataType) {

			JP.Grid.Add(CreatePanel(npdt))
			JP.Grid.ForceRefresh()
			JA.StatusManager.DetectData()

			JC.Notify("New panel created.")
		},
		func(layer *fyne.Container) {
			JA.LayoutManager.SetOverlay(layer)
		},
		func(layer *fyne.Container) {
			JA.LayoutManager.RemoveOverlay(layer)
			JA.StatusManager.SetOverlayShownStatus(false)
		},
	)

	if d != nil {
		d.Show()
	}

}

func OpenPanelEditForm(pk string, uuid string) {

	if JA.StatusManager.IsOverlayShown() {
		return
	}

	JA.StatusManager.SetOverlayShownStatus(true)

	d := JP.NewPanelForm(pk, uuid,
		func(npdt *JT.PanelDataType) {
			SavePanelForm(npdt)
		},
		nil,
		func(layer *fyne.Container) {
			JA.LayoutManager.SetOverlay(layer)
		},
		func(layer *fyne.Container) {
			JA.LayoutManager.RemoveOverlay(layer)
			JA.StatusManager.SetOverlayShownStatus(false)
		})

	if d != nil {
		d.Show()
	}

}

func OpenSettingForm() {

	if JA.StatusManager.IsOverlayShown() {
		return
	}

	JA.StatusManager.SetOverlayShownStatus(true)

	d := JA.NewSettingsForm(
		func() {
			JC.Notify("Saving configuration...")

			go func() {
				if JT.Config.SaveFile() != nil {
					JC.Notify("Configuration saved successfully.")
					JA.StatusManager.DetectData()

					if JT.Config.IsValidTickers() {
						if JT.BT.IsEmpty() {
							JC.Logln("Rebuilding tickers due to empty ticker list")
							JT.TickersInit()

							fyne.Do(func() {
								JX.Grid = JX.NewTickerGrid()
							})
						}

						JA.StatusManager.SetConfigStatus(true)

						JT.TickerCache.SoftReset()
						JC.WorkerManager.Call("update_tickers", JC.CallQueued)

						JT.ExchangeCache.SoftReset()
						JC.WorkerManager.Call("update_rates", JC.CallQueued)
					}
				} else {
					JC.Notify("Failed to save configuration.")
				}
			}()
		},
		func(layer *fyne.Container) {
			JA.LayoutManager.SetOverlay(layer)
		},
		func(layer *fyne.Container) {
			JA.LayoutManager.RemoveOverlay(layer)
			JA.StatusManager.SetOverlayShownStatus(false)
		})

	if d != nil {
		d.Show()
	}
}

func ToggleDraggable() {

	if JA.StatusManager.IsDraggable() {
		JA.StatusManager.DisallowDragging()
	} else {
		JA.StatusManager.AllowDragging()
	}

	JP.Grid.ForceRefresh()
	if JP.ActiveAction != nil {
		JP.ActiveAction.HideTarget()
	}
}

func ScheduledNotificationReset() {
	JC.MainDebouncer.Call("notification_clear", 6000*time.Millisecond, func() {
		nc, ok := JC.NotificationContainer.(*JW.NotificationDisplay)
		if !ok {
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
