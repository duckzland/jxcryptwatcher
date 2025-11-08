package main

import (
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"

	_ "embed"
)

//go:embed static/256x256/jxwatcher.png
var appIconData []byte

//go:embed fonts/Roboto-Regular-subset.ttf
var regularFont []byte

//go:embed fonts/Roboto-Bold-subset.ttf
var boldFont []byte

func updateDisplay() bool {
	const chunkSize = 100

	var allIDs []string
	var chunks [][]string
	var updateCount int
	var mu sync.Mutex

	registered := make(map[string]bool)
	priority := JT.UsePanelMaps().GetVisiblePanels()
	panels := JT.UsePanelMaps().GetData()

	for _, tag := range priority {
		if JT.UsePanelMaps().GetDataByID(tag) == nil {
			continue
		}

		if !registered[tag] {
			allIDs = append(allIDs, tag)
			registered[tag] = true
		}
	}

	for _, pot := range panels {
		id := pot.GetID()
		if !registered[id] {
			allIDs = append(allIDs, id)
			registered[id] = true
		}
	}

	for i := 0; i < len(allIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(allIDs) {
			end = len(allIDs)
		}
		chunks = append(chunks, allIDs[i:end])
	}

	for _, chunk := range chunks {
		ids := chunk
		JC.UseDispatcher().Submit(func() {
			for _, id := range ids {
				pkt := JT.UsePanelMaps().GetDataByID(id)

				if pkt == nil {
					continue
				}

				if pkt.UpdateRate() {

					pkt.UpdateStatus()

					mu.Lock()
					if updateCount == 0 {
						updateCount++
						JC.Notify("Panel display refreshed with latest rates")
					}
					mu.Unlock()
				}
			}

			if updateCount != 0 {
				runtime.GC()
			}
		})
	}

	return true
}

func updateTickerDisplay() bool {

	success := 0
	tickers := []string{}
	if JT.UseConfig().CanDoCMC100() {
		tickers = append(tickers, "cmc100")
	}
	if JT.UseConfig().CanDoFearGreed() {
		tickers = append(tickers, "feargreed")
	}
	if JT.UseConfig().CanDoMarketCap() {
		tickers = append(tickers, "market_cap")
	}
	if JT.UseConfig().CanDoAltSeason() {
		tickers = append(tickers, "altcoin_index")
	}
	if JT.UseConfig().CanDoRSI() {
		tickers = append(tickers, "rsi", "pulse")
	}
	if JT.UseConfig().CanDoETF() {
		tickers = append(tickers, "etf")
	}
	if JT.UseConfig().CanDoDominance() {
		tickers = append(tickers, "dominance")
	}

	for _, key := range tickers {
		tktt := JT.UseTickerMaps().GetDataByType(key)
		for _, tkt := range tktt {
			if tkt.Update() {
				if success == 0 {
					success++
					JC.Notify("Ticker display refreshed with new rates")
				}
			}
		}

		if success != 0 {
			runtime.GC()
		}
	}

	return true
}

func updateRates() bool {

	if JA.UseStatus().IsFetchingRates() {
		return false
	}

	if JT.UsePanelMaps().IsEmpty() {
		return false
	}

	var mu sync.Mutex
	jb := make(map[string]string)
	list := JT.UsePanelMaps().GetData()

	for _, pot := range list {
		pk := JT.UsePanelMaps().GetDataByID(pot.GetID())

		if !JT.UsePanelMaps().ValidatePanel(pk.Get()) {
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

	payloads := make(map[string][]string)

	for _, rk := range jb {
		payloads["rates"] = append(payloads["rates"], rk)
	}
	var hasError int = 0
	successCount := 0

	JC.Notify("Fetching the latest exchange rates...")

	JC.UseFetcher().Dispatch(payloads,
		func(scheduledJobs int) {
			if scheduledJobs > 0 {
				JA.UseStatus().StartFetchingRates()
				JT.UseExchangeCache().SoftReset()
			}
		},
		func(results map[string]JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingRates()

			for _, result := range results {

				ns := detectHTTPResponse(result.Code())
				mu.Lock()
				if hasError == JC.STATUS_SUCCESS || hasError < ns {
					hasError = ns
				}

				if ns == JC.STATUS_SUCCESS {
					successCount++
				}
				mu.Unlock()
			}

			processUpdatePanelComplete(hasError)

			JC.Logf("Exchange Rate updated: %v/%v", successCount, len(jb))

			mu.Lock()
			if successCount != 0 {
				JC.UseWorker().Call("update_display", JC.CallBypassImmediate)
			}
			mu.Unlock()

			JC.UseWorker().Reset("update_rates")
		})

	return true
}

func updateTickers() bool {

	if JA.UseStatus().IsFetchingTickers() {
		return false
	}

	if !JT.UseConfig().IsValidTickers() {
		return false
	}

	var mu sync.Mutex

	// Prepare keys and payloads
	payloads := map[string][]string{}

	if JT.UseConfig().CanDoCMC100() {
		payloads["cmc100"] = []string{"cmc100"}
	}
	if JT.UseConfig().CanDoFearGreed() {
		payloads["feargreed"] = []string{"feargreed"}
	}
	if JT.UseConfig().CanDoMarketCap() {
		payloads["market_cap"] = []string{"market_cap"}
	}
	if JT.UseConfig().CanDoAltSeason() {
		payloads["altcoin_index"] = []string{"altcoin_index"}
	}
	if JT.UseConfig().CanDoRSI() {
		payloads["rsi"] = []string{"rsi"}
	}
	if JT.UseConfig().CanDoETF() {
		payloads["etf"] = []string{"etf"}
	}
	if JT.UseConfig().CanDoDominance() {
		payloads["dominance"] = []string{"dominance"}
	}

	if len(payloads) == 0 {
		return false
	}

	var hasError int = 0
	var successCount int = 0

	JC.Notify("Fetching the latest ticker data...")

	JC.UseFetcher().Dispatch(payloads,
		func(totalJob int) {
			if totalJob > 0 {
				JA.UseStatus().StartFetchingTickers()
				JT.UseTickerCache().SoftReset()
			}
		},
		func(results map[string]JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingTickers()

			for _, result := range results {
				ns := detectHTTPResponse(result.Code())
				switch ns {
				case JC.STATUS_SUCCESS:

					mu.Lock()
					successCount++
					mu.Unlock()
				}

				mu.Lock()
				if hasError == 0 || hasError < ns {
					hasError = ns
				}
				mu.Unlock()
			}

			processUpdateTickerComplete(hasError)

			JC.Logf("Tickers Rate updated: %v/%v", successCount, len(payloads))

			if successCount > 0 {
				updateTickerDisplay()
			}

			JC.UseWorker().Reset("update_tickers")
		})

	return true
}

func detectHTTPResponse(rs int64) int {

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

func processUpdatePanelComplete(status int) {
	switch status {
	case JC.STATUS_SUCCESS:

		JC.Notify("Exchange fetch completed.")
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify("Please check your network connection.")
		JA.UseStatus().SetNetworkStatus(false)

		JT.UsePanelMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.PanelData) bool {
			return pdt.UsePanelKey().IsValueMatchingFloat(0, "<") || pdt.IsStatus(JC.STATE_LOADING)
		})

		fyne.Do(func() {
			JP.UsePanelGrid().UpdatePanelsContent(func(pdt JT.PanelData) bool {
				return true
			})
		})

	case JC.STATUS_CONFIG_ERROR:

		JC.Notify("Please check your settings.")
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(false)

		JT.UsePanelMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.PanelData) bool {
			return pdt.UsePanelKey().IsValueMatchingFloat(0, "<") || pdt.IsStatus(JC.STATE_LOADING)
		})

		fyne.Do(func() {
			JP.UsePanelGrid().UpdatePanelsContent(func(pdt JT.PanelData) bool {
				return true
			})
		})

	case JC.STATUS_BAD_DATA_RECEIVED:

		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)
	}
}

func processUpdateTickerComplete(status int) {

	switch status {
	case JC.STATUS_SUCCESS:

		JC.Notify("Ticker fetch completed.")
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify("Please check your network connection.")
		JA.UseStatus().SetNetworkStatus(false)
		JA.UseStatus().SetConfigStatus(true)

		JT.UseTickerMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.TickerData) bool {
			return !pdt.HasData() || pdt.IsStatus(JC.STATE_LOADING)
		})

		fyne.Do(func() {
			JX.UseTickerGrid().UpdateTickersContent(func(pdt JT.TickerData) bool {
				return true
			})
		})

	case JC.STATUS_CONFIG_ERROR:

		JC.Notify("Please check your settings.")

		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(false)

		JT.UseTickerMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.TickerData) bool {
			return !pdt.HasData() || pdt.IsStatus(JC.STATE_LOADING)
		})

		JX.UseTickerGrid().UpdateTickersContent(func(pdt JT.TickerData) bool {
			return true
		})

	case JC.STATUS_BAD_DATA_RECEIVED:

		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)
	}
}

func processFetchingCryptosComplete(status int) {

	switch status {
	case JC.STATUS_SUCCESS:

		JT.CryptosLoaderInit()
		JA.UseStatus().DetectData()

		if !JA.UseStatus().ValidCryptos() {
			JC.Notify("Failed to convert crypto data to map")
			JA.UseStatus().SetCryptoStatus(false)

			return
		}

		JC.Notify("Crypto map regenerated successfully")

		if JT.UsePanelMaps().RefreshData() {
			fyne.Do(func() {
				JP.UsePanelGrid().ForceRefresh()
			})

			JT.UseExchangeCache().SoftReset()
			JC.UseWorker().Call("update_rates", JC.CallQueued)

			JT.UseTickerCache().SoftReset()
			JC.UseWorker().Call("update_tickers", JC.CallQueued)

			JA.UseStatus().SetCryptoStatus(true)
			JA.UseStatus().SetConfigStatus(true)
			JA.UseStatus().SetNetworkStatus(true)
		}

	case JC.STATUS_NETWORK_ERROR:
		JC.Notify("Please check your network connection.")
		JA.UseStatus().SetNetworkStatus(false)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_CONFIG_ERROR:
		JC.Notify("Please check your settings.")
		JA.UseStatus().SetConfigStatus(false)
		JA.UseStatus().SetNetworkStatus(true)

	case JC.STATUS_BAD_DATA_RECEIVED:
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)
	}
}

func validateRatesCache() bool {

	list := JT.UsePanelMaps().GetData()
	for _, pot := range list {

		// Always get linked data! do not use the copied
		pkt := JT.UsePanelMaps().GetDataByID(pot.GetID())
		pks := pkt.UsePanelKey()
		ck := JT.UseExchangeCache().CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

		if !JT.UseExchangeCache().Has(ck) {
			return false
		}
	}

	return true
}

func validateRateCache(pot JT.PanelData) bool {

	// Always get linked data! do not use the copied
	pkt := JT.UsePanelMaps().GetDataByID(pot.GetID())
	pks := pkt.UsePanelKey()
	ck := JT.UseExchangeCache().CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

	if !JT.UseExchangeCache().Has(ck) {
		return false
	}

	return true
}

func removePanel(uuid string) {

	if JP.UsePanelGrid().RemoveByID(uuid) {
		JC.Logf("Removing panel %s", uuid)

		if JT.UsePanelMaps().Remove(uuid) {
			JP.UsePanelGrid().ForceRefresh()

			JA.UseLayout().RefreshLayout()

			// Prevent UX locking
			go func() {
				if JT.SavePanels() {
					JC.Notify("Panel removed successfully.")
				}
			}()
		}
	}

	JA.UseStatus().DetectData()
}

func savePanelForm(pdt JT.PanelData) {

	JC.Notify("Saving panel settings...")

	JP.UsePanelGrid().ForceRefresh()

	if !JT.UsePanelMaps().ValidatePanel(pdt.Get()) {
		pdt.SetStatus(JC.STATE_BAD_CONFIG)
	}

	// Prevent UX locking
	go func() {

		hasCache := validateRateCache(pdt)

		if hasCache && !pdt.IsStatus(JC.STATE_BAD_CONFIG) {
			if pdt.UpdateRate() {
				pdt.UpdateStatus()
			}
		}

		if JT.SavePanels() {

			if pdt.IsStatus(JC.STATE_BAD_CONFIG) {
				return
			}

			// Only fetch new rates if no cache exists!
			if !hasCache {

				// Force refresh without fail!
				pkt := pdt.UsePanelKey()

				sid := pkt.GetSourceCoinString()
				tid := pkt.GetTargetCoinString()

				payloads := map[string][]string{}
				payloads["rates"] = []string{sid + "|" + tid}

				JC.UseFetcher().Dispatch(payloads,
					func(totalScheduled int) {
					},
					func(results map[string]JC.FetchResultInterface) {
						for _, result := range results {
							JT.UseExchangeCache().SoftReset()

							status := detectHTTPResponse(result.Code())

							switch status {
							case JC.STATUS_SUCCESS:
								if pdt.UpdateRate() {
									pdt.UpdateStatus()
								}

							case JC.STATUS_NETWORK_ERROR, JC.STATUS_CONFIG_ERROR, JC.STATUS_BAD_DATA_RECEIVED:
								pdt.SetStatus(JC.STATE_ERROR)
							}

							processUpdatePanelComplete(status)
						}
					})
			}

			JC.Notify("Panel settings saved.")

		} else {
			JC.Notify("Failed to save panel settings.")
		}
	}()

}

func openNewPanelForm() {
	if JA.UseStatus().IsOverlayShown() {
		return
	}

	JA.UseStatus().SetOverlayShownStatus(true)

	d := JP.NewPanelForm(
		"new",
		"",
		func(npdt JT.PanelData) {
			savePanelForm(npdt)
		},
		func(npdt JT.PanelData) {

			JP.UsePanelGrid().Add(createPanel(npdt))
			JP.UsePanelGrid().ForceRefresh()
			JA.UseStatus().DetectData()

			JC.Notify("New panel created.")
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RegisterOverlay(layer)
			if JC.IsMobile {
				JA.UseStatus().Pause()
				JC.UseDispatcher().Pause()
			}
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RemoveOverlay(layer)
			JA.UseStatus().SetOverlayShownStatus(false)
			if JC.IsMobile {
				JA.UseStatus().Resume()
				JC.UseDispatcher().Resume()
			}
		},
	)

	if d != nil {
		d.Show()
	}

}

func openPanelEditForm(pk string, uuid string) {

	if JA.UseStatus().IsOverlayShown() {
		return
	}

	JA.UseStatus().SetOverlayShownStatus(true)

	d := JP.NewPanelForm(pk, uuid,
		func(npdt JT.PanelData) {
			savePanelForm(npdt)
		},
		nil,
		func(layer *fyne.Container) {
			JA.UseLayout().RegisterOverlay(layer)
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RemoveOverlay(layer)
			JA.UseStatus().SetOverlayShownStatus(false)
		})

	if d != nil {
		d.Show()
	}

}

func openSettingForm() {

	if JA.UseStatus().IsOverlayShown() {
		return
	}

	JA.UseStatus().SetOverlayShownStatus(true)

	d := JA.NewSettingsForm(
		func() {
			JC.Notify("Saving configuration...")

			go func() {
				if JT.ConfigSave() {
					JC.Notify("Configuration saved successfully.")
					JA.UseStatus().DetectData()

					if JT.UseConfig().IsValidTickers() {
						if JT.UseTickerMaps().IsEmpty() {
							JC.Logln("Rebuilding tickers due to empty ticker list")
							JT.TickersInit()

							JX.RegisterTickerGrid()
						}

						JC.UseWorker().Reload()

						JA.UseStatus().SetConfigStatus(true)

						JT.UseTickerCache().SoftReset()
						JC.UseWorker().Call("update_tickers", JC.CallQueued)

						JT.UseExchangeCache().SoftReset()
						JC.UseWorker().Call("update_rates", JC.CallQueued)
					}
				} else {
					JC.Notify("Failed to save configuration.")
				}
			}()
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RegisterOverlay(layer)
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RemoveOverlay(layer)
			JA.UseStatus().SetOverlayShownStatus(false)
		})

	if d != nil {
		d.Show()
	}
}

func toggleDraggable() {

	if JA.UseStatus().IsDraggable() {
		JA.UseStatus().DisallowDragging()
	} else {
		JA.UseStatus().AllowDragging()
	}

	JP.UsePanelGrid().ForceRefresh()
	if JP.UsePanelGrid().HasActiveAction() {
		JP.UsePanelGrid().GetActiveAction().HideTarget()
	}
}

func scheduledNotificationReset() {
	JC.UseDebouncer().Call("notification_clear", 6000*time.Millisecond, func() {

		// Break loop once notification is empty
		if JW.UseNotification().GetText() == "" {
			return
		}

		if JA.UseStatus().IsPaused() {
			return
		}

		// Ensure message shown for at least 6 seconds
		last := JC.UseWorker().GetLastUpdate("notification")
		if time.Since(last) > 6*time.Second {
			JC.Logln("Clearing notification display due to inactivity")
			JW.UseNotification().ClearText()

		} else {
			scheduledNotificationReset()
		}
	})
}

func setAppIcon() {
	icon := fyne.NewStaticResource("jxwatcher.png", appIconData)
	JC.Window.SetIcon(icon)
}

func createPanel(pkt JT.PanelData) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, openPanelEditForm, removePanel)
}

func registerFonts() {
	JC.UseTheme().SetFonts(fyne.TextStyle{Bold: false}, fyne.NewStaticResource("Roboto-Regular.ttf", regularFont))
	JC.UseTheme().SetFonts(fyne.TextStyle{Bold: true}, fyne.NewStaticResource("Roboto-Bold.ttf", boldFont))
}
