package main

import (
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	JN "jxwatcher/animations"
	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

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
						JC.Notify(JC.NotifyPanelDisplayRefreshedWithLatestRates)
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
		tickers = append(tickers, JT.TickerTypeCMC100)
	}
	if JT.UseConfig().CanDoFearGreed() {
		tickers = append(tickers, JT.TickerTypeFearGreed)
	}
	if JT.UseConfig().CanDoMarketCap() {
		tickers = append(tickers, JT.TickerTypeMarketCap)
	}
	if JT.UseConfig().CanDoAltSeason() {
		tickers = append(tickers, JT.TickerTypeAltcoinIndex)
	}
	if JT.UseConfig().CanDoRSI() {
		tickers = append(tickers, JT.TickerTypeRSI, JT.TickerTypePulse)
	}
	if JT.UseConfig().CanDoETF() {
		tickers = append(tickers, JT.TickerTypeETF)
	}
	if JT.UseConfig().CanDoDominance() {
		tickers = append(tickers, JT.TickerTypeDominance)
	}

	for _, key := range tickers {
		tktt := JT.UseTickerMaps().GetDataByType(key)
		for _, tkt := range tktt {
			if tkt.Update() {
				if success == 0 {
					success++
					JC.Notify(JC.NotifyTickerDisplayRefreshedWithNewRates)
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
			jb[sid] = sid + JC.STRING_PIPE + tid
		} else {
			jb[sid] += "," + tid
		}
	}

	if len(jb) == 0 {
		return false
	}

	payloads := make(map[string][]string)

	for _, rk := range jb {
		payloads[JC.ACT_EXCHANGE_GET_RATES] = append(payloads[JC.ACT_EXCHANGE_GET_RATES], rk)
	}
	var hasError int = 0
	successCount := 0

	JC.Notify(JC.NotifyFetchingTheLatestExchangeRates)

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
				JC.UseWorker().Call(JC.ACT_PANEL_UPDATE, JC.CallBypassImmediate)
			}
			mu.Unlock()

			JC.UseWorker().Reset(JC.ACT_EXCHANGE_UPDATE_RATES)
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
		payloads[JT.TickerTypeCMC100] = []string{JT.TickerTypeCMC100}
	}
	if JT.UseConfig().CanDoFearGreed() {
		payloads[JT.TickerTypeFearGreed] = []string{JT.TickerTypeFearGreed}
	}
	if JT.UseConfig().CanDoMarketCap() {
		payloads[JT.TickerTypeMarketCap] = []string{JT.TickerTypeMarketCap}
	}
	if JT.UseConfig().CanDoAltSeason() {
		payloads[JT.TickerTypeAltcoinIndex] = []string{JT.TickerTypeAltcoinIndex}
	}
	if JT.UseConfig().CanDoRSI() {
		payloads[JT.TickerTypeRSI] = []string{JT.TickerTypeRSI}
	}
	if JT.UseConfig().CanDoETF() {
		payloads[JT.TickerTypeETF] = []string{JT.TickerTypeETF}
	}
	if JT.UseConfig().CanDoDominance() {
		payloads[JT.TickerTypeDominance] = []string{JT.TickerTypeDominance}
	}

	if len(payloads) == 0 {
		return false
	}

	var hasError int = 0
	var successCount int = 0

	JC.Notify(JC.NotifyFetchingTheLatestTickerData)

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

			JC.UseWorker().Reset(JC.ACT_TICKER_UPDATE)
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

		JC.Notify(JC.NotifyExchangeFetchCompleted)
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify(JC.NotifyPleaseCheckYourNetworkConnection)
		JA.UseStatus().SetNetworkStatus(false)

		JT.UsePanelMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.PanelData) bool {
			return pdt.UsePanelKey().IsValueMatchingFloat(0, JC.STRING_LESS) || pdt.IsStatus(JC.STATE_LOADING)
		})

		fyne.Do(func() {
			JP.UsePanelGrid().UpdatePanelsContent(func(pdt JT.PanelData) bool {
				return true
			})
		})

	case JC.STATUS_CONFIG_ERROR:

		JC.Notify(JC.NotifyPleaseCheckYourSettings)
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(false)

		JT.UsePanelMaps().ChangeStatus(JC.STATE_ERROR, func(pdt JT.PanelData) bool {
			return pdt.UsePanelKey().IsValueMatchingFloat(0, JC.STRING_LESS) || pdt.IsStatus(JC.STATE_LOADING)
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

		JC.Notify(JC.NotifyTickerFetchCompleted)
		JA.UseStatus().SetNetworkStatus(true)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_NETWORK_ERROR:

		JC.Notify(JC.NotifyPleaseCheckYourNetworkConnection)
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

		JC.Notify(JC.NotifyPleaseCheckYourSettings)

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
			JC.Notify(JC.NotifyFailedToConvertCryptoDataToMap)
			JA.UseStatus().SetCryptoStatus(false)

			return
		}

		JC.Notify(JC.NotifyCryptoMapRegeneratedSuccessfully)

		if JT.UsePanelMaps().RefreshData() {
			fyne.Do(func() {
				JP.UsePanelGrid().ForceRefresh()
			})

			JT.UseExchangeCache().SoftReset()
			JC.UseWorker().Call(JC.ACT_EXCHANGE_UPDATE_RATES, JC.CallQueued)

			JT.UseTickerCache().SoftReset()
			JC.UseWorker().Call(JC.ACT_TICKER_UPDATE, JC.CallQueued)

			JA.UseStatus().SetCryptoStatus(true)
			JA.UseStatus().SetConfigStatus(true)
			JA.UseStatus().SetNetworkStatus(true)
		}

	case JC.STATUS_NETWORK_ERROR:
		JC.Notify(JC.NotifyPleaseCheckYourNetworkConnection)
		JA.UseStatus().SetNetworkStatus(false)
		JA.UseStatus().SetConfigStatus(true)

	case JC.STATUS_CONFIG_ERROR:
		JC.Notify(JC.NotifyPleaseCheckYourSettings)
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
					JC.Notify(JC.NotifyPanelRemovedSuccessfully)
				}
			}()
		}
	}

	JA.UseStatus().DetectData()
}

func savePanelForm(pdt JT.PanelData) {

	JC.Notify(JC.NotifySavingPanelSettings)

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
				payloads[JC.ACT_EXCHANGE_GET_RATES] = []string{sid + JC.STRING_PIPE + tid}

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

			JC.Notify(JC.NotifyPanelSettingsSaved)

		} else {
			JC.Notify(JC.NotifyFailedToSavePanelSettings)
		}
	}()

}

func openNewPanelForm() {
	if JA.UseStatus().IsOverlayShown() {
		return
	}

	JA.UseStatus().SetOverlayShownStatus(true)

	d := JP.NewPanelForm(
		JC.ACT_PANEL_NEW,
		JC.STRING_EMPTY,
		func(npdt JT.PanelData) {
			savePanelForm(npdt)
		},
		func(npdt JT.PanelData) {

			JP.UsePanelGrid().Add(createPanel(npdt))
			JP.UsePanelGrid().ForceRefresh()
			JA.UseStatus().DetectData()

			JC.Notify(JC.NotifyNewPanelCreated)
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
			JC.Notify(JC.NotifySavingConfiguration)

			go func() {
				if JT.ConfigSave() {
					JC.Notify(JC.NotifyConfigurationSavedSuccessfully)
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
						JC.UseWorker().Call(JC.ACT_TICKER_UPDATE, JC.CallQueued)

						JT.UseExchangeCache().SoftReset()
						JC.UseWorker().Call(JC.ACT_EXCHANGE_UPDATE_RATES, JC.CallQueued)
					}
				} else {
					JC.Notify(JC.NotifyFailedToSaveConfiguration)
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
	JC.UseDebouncer().Call(JC.ACT_NOTIFICATION_CLEAR, 6000*time.Millisecond, func() {

		// Break loop once notification is empty
		if JW.UseNotification().GetText() == JC.STRING_EMPTY {
			return
		}

		if JA.UseStatus().IsPaused() {
			return
		}

		// Ensure message shown for at least 6 seconds
		last := JC.UseWorker().GetLastUpdate(JC.ACT_NOTIFICATION_PUSH)
		if time.Since(last) > 6*time.Second {
			JC.Logln("Clearing notification display due to inactivity")
			JW.UseNotification().ClearText()

		} else {
			scheduledNotificationReset()
		}
	})
}

func createPanel(pkt JT.PanelData) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, openPanelEditForm, removePanel)
}

func appShutdown() {
	worker := JC.UseWorker()
	if worker != nil {
		worker.Destroy()
	}

	fetcher := JC.UseFetcher()
	if fetcher != nil {
		fetcher.Destroy()
	}

	debouncer := JC.UseDebouncer()
	if debouncer != nil {
		debouncer.Destroy()
	}

	dispatcher := JC.UseDispatcher()
	if dispatcher != nil {
		dispatcher.Destroy()
	}

	animDispatcher := JN.UseAnimationDispatcher()
	if animDispatcher != nil {
		animDispatcher.Destroy()
	}

	status := JA.UseStatus()
	if status == nil {
		JC.Logln("Refused to take snapshot as status is nil")
		return
	}
	if !status.IsReady() {
		JC.Logln("Refused to take snapshot as app is not ready yet")
		return
	}

	snapshot := JA.UseSnapshot()
	if snapshot != nil && !snapshot.IsSnapshotted() {
		snapshot.Save()
	}

}
