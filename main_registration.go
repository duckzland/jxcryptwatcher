package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func RegisterActions() {

	// Refresh ticker data
	JA.AppActions.AddButton(JW.NewActionButton("refresh_cryptos", "", theme.ViewRestoreIcon(), "Refresh cryptos data", "disabled",
		func(btn *JW.ActionButton) {
			JC.FetcherManager.Call("cryptos_map", nil)
		},
		func(btn *JW.ActionButton) {
			if !JA.AppStatus.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.AppStatus.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidConfig() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingRates() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingCryptos() {
				btn.Progress()
				return
			}

			if !JA.AppStatus.ValidCryptos() {
				btn.Error()
				return
			}

			if !JA.AppStatus.IsValidCrypto() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Refresh exchange rates
	JA.AppActions.AddButton(JW.NewActionButton("refresh_rates", "", theme.ViewRefreshIcon(), "Update rates from exchange", "disabled",
		func(btn *JW.ActionButton) {
			// Open the network status temporarily
			JA.AppStatus.SetNetworkStatus(true)

			// Force update
			JT.ExchangeCache.SoftReset()
			JC.WorkerManager.Call("update_rates", JC.CallDebounced)

			// Force update
			JT.TickerCache.SoftReset()
			JC.WorkerManager.Call("update_tickers", JC.CallDebounced)

		},
		func(btn *JW.ActionButton) {
			if !JA.AppStatus.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.AppStatus.IsDraggable() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidConfig() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidCryptos() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidPanels() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.IsValidConfig() {
				btn.Error()
				return
			}

			if !JA.AppStatus.IsGoodNetworkStatus() {
				btn.Error()
				return
			}

			if JA.AppStatus.IsFetchingRates() {
				btn.Progress()
				return
			}

			if JA.AppStatus.IsFetchingTickers() {
				btn.Progress()
				return
			}

			btn.Enable()
		}))

	// Open settings
	JA.AppActions.AddButton(JW.NewActionButton("open_settings", "", theme.SettingsIcon(), "Open settings", "disabled",
		func(btn *JW.ActionButton) {
			OpenSettingForm()
		},
		func(btn *JW.ActionButton) {
			if !JA.AppStatus.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.AppStatus.IsFetchingTickers() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingRates() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidConfig() {
				btn.Error()
				return
			}

			if !JA.AppStatus.IsValidConfig() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Panel drag toggle
	JA.AppActions.AddButton(JW.NewActionButton("toggle_drag", "", theme.ContentPasteIcon(), "Enable Reordering", "disabled",
		func(btn *JW.ActionButton) {
			ToggleDraggable()
		},
		func(btn *JW.ActionButton) {
			if !JA.AppStatus.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.AppStatus.IsFetchingCryptos() {
				JA.AppStatus.DisallowDragging()
				btn.Disable()
				return
			}

			if JA.AppStatus.IsFetchingRates() {
				JA.AppStatus.DisallowDragging()
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidPanels() {
				JA.AppStatus.DisallowDragging()
				btn.Disable()
				return
			}

			if JT.BP.TotalData() < 2 {
				JA.AppStatus.DisallowDragging()
				btn.Disable()
				return
			}

			if JA.AppStatus.IsDraggable() {
				btn.Active()
				return
			}

			btn.Enable()
		}))

	// Add new panel
	JA.AppActions.AddButton(JW.NewActionButton("add_panel", "", theme.ContentAddIcon(), "Add new panel", "disabled",
		func(btn *JW.ActionButton) {
			OpenNewPanelForm()
		},
		func(btn *JW.ActionButton) {
			if !JA.AppStatus.IsReady() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.AppStatus.IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if JA.AppStatus.IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.AppStatus.ValidCryptos() {
				btn.Disable()
				return
			}

			btn.Enable()
		}))
}

func RegisterWorkers() {

	JC.WorkerManager.RegisterSleeper("update_display", func() {
		if UpdateDisplay() {
			JC.UpdateDisplayTimestamp = time.Now()
		}
	}, 200, func() bool {

		if !JA.AppStatus.ValidConfig() {
			JC.Logln("Unable to refresh display: invalid configuration")
			return false
		}

		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to refresh display: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
			JC.Logln("Unable to refresh display: app is paused")
			return false
		}

		if !JT.ExchangeCache.HasData() {
			JC.Notify("Unable to refresh display: no cached data")
			return false
		}

		if !JT.ExchangeCache.GetTimestamp().After(JC.UpdateDisplayTimestamp) {
			JC.Notify("Unable to refresh display: Data is newer than display timestamp")
			return false
		}

		if !JA.AppStatus.ValidPanels() {
			JC.Logln("Unable to refresh display: No valid panels configured")
			return false
		}

		return true
	})

	JC.WorkerManager.Register("update_rates", func() {
		UpdateRates()
	}, func() int64 {
		return max(JT.Config.Delay*1000, 30000)
	}, 1000, func() bool {

		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to refresh rates: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
			JC.Logln("Unable to refresh rates: app is paused")
			return false
		}

		if JA.AppStatus.IsDraggable() {
			JC.Logln("Unable to refresh rates: app is dragging")
			return false
		}

		if !JA.AppStatus.ValidConfig() {
			JC.Notify("Unable to refresh rates: invalid configuration")
			return false
		}

		if !JT.ExchangeCache.ShouldRefresh() {
			JC.Logln("Unable to refresh rates: not cleared should refresh yet")
			return false
		}

		if !JA.AppStatus.ValidPanels() {
			JC.Logln("Unable to refresh rates: No valid panels configured")
			return false
		}

		return true
	})

	JC.WorkerManager.Register("update_tickers", func() {
		UpdateTickers()
	}, func() int64 {
		return max(JT.Config.Delay*1000, 30000)
	}, 5000, func() bool {

		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to refresh tickers: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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

		return true
	})

	JC.WorkerManager.RegisterBuffered("notification", func(messages []string) bool {
		if len(messages) == 0 {
			return false
		}
		latest := messages[len(messages)-1]

		nc, ok := JC.NotificationContainer.(*JW.NotificationDisplay)
		if !ok {
			return false
		}

		fyne.Do(func() {
			nc.UpdateText(latest)
		})

		ScheduledNotificationReset()

		return true

	}, 1000, 100, 1000, 2000, func() bool {
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to do notification: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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

		defer JA.AppStatus.EndFetchingCryptos()

		status := DetectHTTPResponse(result.Code)
		ProcessFetchingCryptosComplete(status)

	}, func() bool {

		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch cryptos: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
			JC.Logln("Unable to fetch cryptos: app is paused")
			return false
		}

		if !JA.AppStatus.ValidConfig() {
			JC.Notify("Invalid configuration. Unable to reset cryptos map.")
			JC.Logln("Unable to do fetch cryptos: Invalid config")
			return false
		}

		if JA.AppStatus.IsFetchingCryptos() {
			JC.Logln("Unable to do fetch cryptos: Another fetcher is running")
			return false
		}

		JA.AppStatus.StartFetchingCryptos()

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
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch CMC100: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch fear greed: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch marketcap: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch altcoin index: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
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
		if !JA.AppStatus.IsReady() {
			JC.Logln("Unable to fetch rates: app is not ready yet")
			return false
		}

		if JA.AppStatus.IsPaused() {
			JC.Logln("Unable to fetch rates: app is paused")
			return false
		}

		if !JA.AppStatus.ValidPanels() {
			JC.Logln("Unable to rates: no configured panels")
			return false
		}

		return true
	})
}

func RegisterLifecycle() {

	// Hook into lifecycle events
	if lc := JC.App.Lifecycle(); lc != nil {

		var snapshotSaved bool = false

		lc.SetOnEnteredForeground(func() {
			JC.Logln("App entered foreground")

			snapshotSaved = false

			if JC.IsMobile {
				JC.Logln("Battery Saver: Continuing apps")
				JC.WorkerManager.ResumeAll()
				JA.AppStatus.ContinueApp()
			}

			if !JA.AppStatus.IsReady() {
				JC.Logln("Refused to fetch data as app is not ready yet")
				return
			}

			if !JA.AppStatus.HasError() && JC.IsMobile {
				// Force Refresh
				JT.ExchangeCache.SoftReset()
				JC.WorkerManager.Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.TickerCache.SoftReset()
				JC.WorkerManager.Call("update_tickers", JC.CallImmediate)
			}
		})
		lc.SetOnExitedForeground(func() {
			JC.Logln("App exited foreground")

			if JC.IsMobile {
				JC.Logln("Battery Saver: Pausing apps")
				JA.AppStatus.PauseApp()
				JC.WorkerManager.PauseAll()
			}

			if !JA.AppStatus.IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved && JC.IsMobile {
				JA.AppSnapshot.Save()
				snapshotSaved = true
			}
		})
		lc.SetOnStopped(func() {
			JC.Logln("App stopped")

			if !JA.AppStatus.IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved {
				JA.AppSnapshot.Save()
				snapshotSaved = true
			}
		})
	}
}

func RegisterDispatcher() {
	if JC.IsMobile {
		JC.AnimDispatcher = JC.NewDispatcher(10000, 1, 48*time.Millisecond)
	} else {
		JC.AnimDispatcher = JC.NewDispatcher(10000, 10, 1*time.Millisecond)
	}
}
