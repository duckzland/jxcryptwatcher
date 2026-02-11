package main

import (
	"runtime"
	"strings"
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

	if JC.IsShuttingDown() {
		return false
	}

	if !JA.UseStatus().IsReady() {
		JC.Logln("Unable to refresh display: app is not ready yet")
		return false
	}

	if JA.UseStatus().IsPaused() {
		JC.Logln("Unable to refresh display: app is paused")
		return false
	}

	var allIDs []string

	recentUpdates := JT.UseExchangeCache().GetRecentUpdates()
	registered := make(map[string]bool)
	priority := JT.UsePanelMaps().GetVisiblePanels()
	panels := JT.UsePanelMaps().GetData()

	if panels == nil || len(panels) == 0 {
		JC.Logln("Unable to refresh display: No panels available to update")
		return false
	}

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
		pkt := pot.UsePanelKey()
		ck := JT.UseExchangeCache().CreateKeyFromInt(pkt.GetSourceCoinInt(), pkt.GetTargetCoinInt())

		if !registered[id] {
			_, hasRecentUpdates := recentUpdates[ck]

			// Edge case where panel stuck not in loaded state, refresh status and rate using exchangerate directly
			if !hasRecentUpdates && !pot.IsStatus(JC.STATE_LOADED) {
				pot.UpdateRate()
				pot.UpdateStatus()
				continue
			}

			allIDs = append(allIDs, id)
			registered[id] = true
		}
	}

	ttl := len(allIDs)

	if ttl == 0 {
		JC.Logln("Unable to refresh display: No panels eligible for update")
		return false
	}

	updateCount := 0

	for _, id := range allIDs {
		if JC.IsShuttingDown() {
			return false
		}

		pn := JT.UsePanelMaps().GetDataByID(id)
		if pn == nil {
			continue
		}

		pkt := pn.UsePanelKey()
		if pkt == nil {
			continue
		}

		ck := JT.UseExchangeCache().CreateKeyFromInt(pkt.GetSourceCoinInt(), pkt.GetTargetCoinInt())
		val, ok := recentUpdates[ck]
		if !ok || val == nil {
			continue
		}

		if pn.SetRate(val) {

			pn.UpdateStatus()
			if updateCount == 0 {
				JC.Notify(JC.NotifyPanelDisplayRefreshedWithLatestRates)
			}

			updateCount++
		}
	}

	JA.UseLayout().RegisterDisplayUpdate(time.Now())

	JC.Logf("Panels display updated: %d/%d/%d/%d", len(recentUpdates), updateCount, len(allIDs), len(panels))

	runtime.GC()

	return true
}

func updateTickerDisplay() bool {

	if JC.IsShuttingDown() {
		return false
	}

	if !JA.UseStatus().IsTickerShown() {
		return false
	}

	recentUpdates := JT.UseTickerCache().GetRecentUpdates()
	if recentUpdates == nil || len(recentUpdates) == 0 {
		JC.Logln("Unable to refresh tickers: No recent ticker rates update available")
		return false
	}

	success := 0
	tickers := make([]string, 0, 8)

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

	if len(tickers) == 0 {
		JC.Logln("Unable to refresh tickers: No configured tickers")
		return false
	}

	for _, key := range tickers {
		rate, hasRateUpdate := recentUpdates[key]
		tktt := JT.UseTickerMaps().GetDataByType(key)
		for _, tkt := range tktt {

			// Edge case where ticker is stuck not at loaded state, try to update via tickercache data
			if !hasRateUpdate {
				if !tkt.IsStatus(JC.STATE_LOADED) {
					tkt.Update()
				}
				continue
			}

			if JC.IsShuttingDown() {
				return false
			}

			tkt.Insert(rate)
			tkt.UpdateStatus()

			success++
		}
	}

	if success != 0 {
		if JA.UseStatus().IsTickerShown() {
			JC.Notify(JC.NotifyTickerDisplayRefreshedWithNewRates)
		}
	}

	JC.Logf("Tickers display updated: %d/%d/%d", len(recentUpdates), success, len(tickers))

	runtime.GC()

	return true
}

func updateRates() bool {

	if JC.IsShuttingDown() {
		return false
	}

	if JA.UseStatus().IsFetchingRates() {
		return false
	}

	if JT.UsePanelMaps().IsEmpty() {
		return false
	}

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
		JC.Logln("Unable to retrieve rates: No valid payload generated")
		return false
	}

	for sid, val := range jb {
		parts := strings.Split(strings.TrimPrefix(val, sid+JC.STRING_PIPE), ",")
		seen := make(map[string]struct{})
		var uniq []string
		for _, p := range parts {
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
				uniq = append(uniq, p)
			}
		}
		jb[sid] = sid + JC.STRING_PIPE + strings.Join(uniq, ",")
	}

	payloads := make(map[string][]string, len(jb))

	for _, rk := range jb {
		payloads[JC.ACT_EXCHANGE_GET_RATES] = append(payloads[JC.ACT_EXCHANGE_GET_RATES], rk)
	}

	JC.Notify(JC.NotifyFetchingTheLatestExchangeRates)

	JC.UseFetcher().Call(payloads,
		func(scheduledJobs int) {

			if JC.IsShuttingDown() {
				return
			}

			if scheduledJobs > 0 {
				JA.UseStatus().StartFetchingRates()
				JT.UseExchangeCache().SoftReset()
			}
		},
		func(results map[string]JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingRates()
			defer JC.UseWorker().Reset(JC.ACT_EXCHANGE_UPDATE_RATES)

			hasError := 0
			successCount := 0

			for _, result := range results {

				if JC.IsShuttingDown() {
					return
				}

				ns := detectHTTPResponse(result.Code())
				if hasError == JC.STATUS_SUCCESS || hasError < ns {
					hasError = ns
				}

				if ns == JC.STATUS_SUCCESS {
					successCount++
				}
			}

			processUpdatePanelComplete(hasError)

			JC.Logf("Exchange rate updated: %v/%v", successCount, len(jb))

			if successCount != 0 {
				updateDisplay()
			}
		},
		func() {
			JA.UseStatus().EndFetchingRates()
		})

	runtime.GC()

	return true
}

func updateTickers() bool {

	if JC.IsShuttingDown() {
		return false
	}

	if JA.UseStatus().IsFetchingTickers() {
		return false
	}

	if !JT.UseConfig().IsValidTickers() {
		return false
	}

	if !JA.UseStatus().IsTickerShown() {
		return false
	}

	// Prepare keys and payloads
	payloads := make(map[string][]string, 8)

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

	if JA.UseStatus().IsTickerShown() {
		JC.Notify(JC.NotifyFetchingTheLatestTickerData)
	}

	JC.UseFetcher().Call(payloads,
		func(totalJob int) {
			if JC.IsShuttingDown() {
				return
			}

			if totalJob > 0 {
				JA.UseStatus().StartFetchingTickers()
				JT.UseTickerCache().SoftReset()
			}
		},
		func(results map[string]JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingTickers()
			defer JC.UseWorker().Reset(JC.ACT_TICKER_UPDATE)

			hasError := 0
			successCount := 0

			for _, result := range results {
				if JC.IsShuttingDown() {
					return
				}

				ns := detectHTTPResponse(result.Code())
				switch ns {
				case JC.STATUS_SUCCESS:
					successCount++
				}

				if hasError == 0 || hasError < ns {
					hasError = ns
				}
			}

			processUpdateTickerComplete(hasError)

			JC.Logf("Tickers rate updated: %v/%v", successCount, len(payloads))

			if successCount > 0 {
				updateTickerDisplay()
			}
		},
		func() {
			JA.UseStatus().EndFetchingTickers()
		})

	runtime.GC()

	return true
}

func detectHTTPResponse(rs int64) int {

	switch rs {
	case JC.NETWORKING_SUCCESS:
		return JC.STATUS_SUCCESS

	case JC.NETWORKING_ERROR_CONNECTION, JC.NETWORKING_RATE_LIMIT, JC.NETWORKING_ERROR_FIREWALL, JC.NETWORKING_NO_INTERNET:
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
		JA.UseStatus().SetConfigStatus(true)

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

		if JA.UseStatus().IsTickerShown() {
			JC.Notify(JC.NotifyTickerFetchCompleted)
		}
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

		if !JA.UseStatus().IsValidCrypto() {
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

				JC.UseFetcher().Call(payloads,
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
					},
					func() {
						JA.UseStatus().EndFetchingRates()
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
				JN.UseAnimationDispatcher().Pause()
				JA.UseStatus().Pause()
			}
		},
		func(layer *fyne.Container) {
			JA.UseLayout().RemoveOverlay(layer)
			JA.UseStatus().SetOverlayShownStatus(false)
			if JC.IsMobile {
				JN.UseAnimationDispatcher().Resume()
				JA.UseStatus().Resume()
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

func notificationResetCallback() {
	if JW.UseNotification().GetText() == JC.STRING_EMPTY {
		return
	}
	if JA.UseStatus().IsPaused() {
		return
	}
	last := JC.UseWorker().GetLastUpdate(JC.ACT_NOTIFICATION_PUSH)
	if time.Since(last) > 6*time.Second {
		JC.Logln("Clearing notification display due to inactivity")
		JW.UseNotification().ClearText()
	} else {
		scheduledNotificationReset()
	}
}

func scheduledNotificationReset() {
	JC.UseDebouncer().Call(JC.ACT_NOTIFICATION_CLEAR, 6000*time.Millisecond, notificationResetCallback)
}

func createPanel(pkt JT.PanelData) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, openPanelEditForm, removePanel)
}

func appShutdown() {

	// Must snapshot first, at this point the mutex can handle the locking properly!
	status := JA.UseStatus()
	snapshot := JA.UseSnapshot()
	if status != nil && snapshot != nil && !snapshot.IsSnapshotted() && status.IsReady() {
		snapshot.Save()
	}

	JC.ShutdownCancel()

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

	animDispatcher := JN.UseAnimationDispatcher()
	if animDispatcher != nil {
		animDispatcher.Destroy()
	}
}
