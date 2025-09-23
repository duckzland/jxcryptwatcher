package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JA "jxwatcher/apps"
	"jxwatcher/core"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func RegisterActions() {

	// Refresh ticker data
	JA.UseAction().Add(JW.NewActionButton("refresh_cryptos", "", theme.ViewRestoreIcon(), "Refresh cryptos data", "disabled",
		func(btn JW.ActionButton) {
			JC.UseFetcher().Call("cryptos_map", nil)
		},
		func(btn JW.ActionButton) {
			if !JA.UseStatus().IsReady() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidConfig() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingRates() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				btn.Progress()
				return
			}

			if !JA.UseStatus().ValidCryptos() {
				btn.Error()
				return
			}

			if !JA.UseStatus().IsValidCrypto() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Refresh exchange rates
	JA.UseAction().Add(JW.NewActionButton("refresh_rates", "", theme.ViewRefreshIcon(), "Update rates from exchange", "disabled",
		func(btn JW.ActionButton) {
			// Open the network status temporarily
			JA.UseStatus().SetNetworkStatus(true)

			// Force update
			JT.UseExchangeCache().SoftReset()
			JC.UseWorker().Call("update_rates", JC.CallDebounced)

			// Force update
			JT.UseTickerCache().SoftReset()
			JC.UseWorker().Call("update_tickers", JC.CallDebounced)

		},
		func(btn JW.ActionButton) {
			if !JA.UseStatus().IsReady() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsDraggable() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidConfig() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidCryptos() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidPanels() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().IsValidConfig() {
				btn.Error()
				return
			}

			if !JA.UseStatus().IsGoodNetworkStatus() {
				btn.Error()
				return
			}

			if JA.UseStatus().IsFetchingRates() {
				btn.Progress()
				return
			}

			if JA.UseStatus().IsFetchingTickers() {
				btn.Progress()
				return
			}

			btn.Enable()
		}))

	// Open settings
	JA.UseAction().Add(JW.NewActionButton("open_settings", "", theme.SettingsIcon(), "Open settings", "disabled",
		func(btn JW.ActionButton) {
			OpenSettingForm()
		},
		func(btn JW.ActionButton) {
			if !JA.UseStatus().IsReady() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsFetchingTickers() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingRates() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidConfig() {
				btn.Error()
				return
			}

			if !JA.UseStatus().IsValidConfig() {
				btn.Error()
				return
			}

			btn.Enable()
		}))

	// Panel drag toggle
	JA.UseAction().Add(JW.NewActionButton("toggle_drag", "", theme.ContentPasteIcon(), "Enable Reordering", "disabled",
		func(btn JW.ActionButton) {
			ToggleDraggable()
		},
		func(btn JW.ActionButton) {
			if !JA.UseStatus().IsReady() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				JA.UseStatus().DisallowDragging()
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingRates() {
				JA.UseStatus().DisallowDragging()
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidPanels() {
				JA.UseStatus().DisallowDragging()
				btn.Disable()
				return
			}

			if JT.UsePanelMaps().TotalData() < 2 {
				JA.UseStatus().DisallowDragging()
				btn.Disable()
				return
			}

			if JA.UseStatus().IsDraggable() {
				btn.Active()
				return
			}

			btn.Enable()
		}))

	// Add new panel
	JA.UseAction().Add(JW.NewActionButton("add_panel", "", theme.ContentAddIcon(), "Add new panel", "disabled",
		func(btn JW.ActionButton) {
			OpenNewPanelForm()
		},
		func(btn JW.ActionButton) {
			if !JA.UseStatus().IsReady() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsDraggable() {
				btn.Disable()
				return
			}

			if !JA.UseStatus().ValidCryptos() {
				btn.Disable()
				return
			}

			btn.Enable()
		}))
}

func RegisterWorkers() {

	JC.UseWorker().RegisterSleeper("update_display", func() {
		if UpdateDisplay() {
			JA.UseLayout().SetLastDisplayUpdate(time.Now())
		}
	}, 200, func() bool {

		if !JA.UseStatus().ValidConfig() {
			JC.Logln("Unable to refresh display: invalid configuration")
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

		if !JT.UseExchangeCache().HasData() {
			JC.Notify("Unable to refresh display: no cached data")
			return false
		}

		if !JT.UseExchangeCache().GetTimestamp().After(JA.UseLayout().GetLastDisplayUpdate()) {
			JC.Notify("Unable to refresh display: Data is newer than display timestamp")
			return false
		}

		if !JA.UseStatus().ValidPanels() {
			JC.Logln("Unable to refresh display: No valid panels configured")
			return false
		}

		return true
	})

	JC.UseWorker().Register("update_rates", func() {
		UpdateRates()
	}, func() int64 {
		return max(JT.UseConfig().Delay*1000, 30000)
	}, 1000, func() bool {

		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to refresh rates: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to refresh rates: app is paused")
			return false
		}

		if JA.UseStatus().IsDraggable() {
			JC.Logln("Unable to refresh rates: app is dragging")
			return false
		}

		if !JA.UseStatus().ValidConfig() {
			JC.Notify("Unable to refresh rates: invalid configuration")
			return false
		}

		if !JT.UseExchangeCache().ShouldRefresh() {
			JC.Logln("Unable to refresh rates: not cleared should refresh yet")
			return false
		}

		if !JA.UseStatus().ValidPanels() {
			JC.Logln("Unable to refresh rates: No valid panels configured")
			return false
		}

		return true
	})

	JC.UseWorker().Register("update_tickers", func() {
		UpdateTickers()
	}, func() int64 {
		return max(JT.UseConfig().Delay*1000, 30000)
	}, 5000, func() bool {

		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to refresh tickers: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to refresh tickers: app is paused")
			return false
		}

		if !JT.UseConfig().IsValidTickers() {
			JC.Logln("Unable to refresh tickers: Invalid ticker configuration")
			return false
		}

		if !JT.UseTickerCache().ShouldRefresh() {
			JC.Logln("Unable to refresh tickers: Ticker cache shouldn't be refreshed yet")
			return false
		}

		return true
	})

	JC.UseWorker().RegisterBuffered("notification", func(messages []string) bool {
		if len(messages) == 0 {
			return false
		}
		latest := messages[len(messages)-1]
		fyne.Do(func() {
			JW.UseNotification().UpdateText(latest)
		})

		ScheduledNotificationReset()

		return true

	}, 1000, 100, 1000, 2000, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to do notification: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to do notification: app is paused")
			return false
		}

		return true
	})
}

func RegisterFetchers() {

	var delay int64 = 100

	JC.UseFetcher().Register("cryptos_map", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			code := JT.UseCryptosLoader().GetCryptos()

			return JC.FetchResult{
				Code: code,
				Data: JT.UseCryptosLoader(),
			}, nil
		},
	}, 0, func(result JC.FetchResult) {

		defer JA.UseStatus().EndFetchingCryptos()

		status := DetectHTTPResponse(result.Code)
		ProcessFetchingCryptosComplete(status)

	}, func() bool {

		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch cryptos: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch cryptos: app is paused")
			return false
		}

		if !JA.UseStatus().ValidConfig() {
			JC.Notify("Invalid configuration. Unable to reset cryptos map.")
			JC.Logln("Unable to do fetch cryptos: Invalid config")
			return false
		}

		if JA.UseStatus().IsFetchingCryptos() {
			JC.Logln("Unable to do fetch cryptos: Another fetcher is running")
			return false
		}

		JA.UseStatus().StartFetchingCryptos()

		return true
	})

	JC.UseFetcher().Register("cmc100", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.NewCMC100Fetcher()
			code := ft.GetRate()

			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch CMC100: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch CMC100: app is paused")
			return false
		}

		if !JT.UseConfig().CanDoCMC100() {
			JC.Logln("Unable to fetch CMC100: Invalid config")
			return false
		}

		return true
	})

	JC.UseFetcher().Register("feargreed", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.NewFearGreedFetcher()
			code := ft.GetRate()

			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch fear greed: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch fear greed app is paused")
			return false
		}

		if !JT.UseConfig().CanDoFearGreed() {
			JC.Logln("Unable to fetch fear greed: Invalid config")
			return false
		}

		return true
	})

	JC.UseFetcher().Register("market_cap", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.NewMarketCapFetcher()
			code := ft.GetRate()

			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch marketcap: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch marketcap: app is paused")
			return false
		}

		if !JT.UseConfig().CanDoMarketCap() {
			JC.Logln("Unable to fetch marketcap: Invalid config")
			return false
		}

		return true
	})

	JC.UseFetcher().Register("altcoin_index", &JC.GenericFetcher{
		Handler: func(ctx context.Context) (JC.FetchResult, error) {
			ft := JT.NewAltSeasonFetcher()
			code := ft.GetRate()

			return JC.FetchResult{Code: code}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch altcoin index: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch altcoin index: app is paused")
			return false
		}

		if !JT.UseConfig().CanDoAltSeason() {
			JC.Logln("Unable to fetch altcoin index: Invalid config")
			return false
		}

		return true
	})

	JC.UseFetcher().Register("rates", &JC.DynamicPayloadFetcher{
		Handler: func(ctx context.Context, payload any) (JC.FetchResult, error) {
			rk, ok := payload.(string)

			if !ok {
				return JC.FetchResult{Code: JC.NETWORKING_BAD_PAYLOAD}, fmt.Errorf("invalid rk")
			}

			ex := JT.NewExchangeResults()
			code := ex.GetRate(rk)
			return JC.FetchResult{Code: code, Data: ex}, nil
		},
	}, delay, func(fr JC.FetchResult) {
		// Process results?

	}, func() bool {
		if !JA.UseStatus().IsReady() {
			JC.Logln("Unable to fetch rates: app is not ready yet")
			return false
		}

		if JA.UseStatus().IsPaused() {
			JC.Logln("Unable to fetch rates: app is paused")
			return false
		}

		if !JA.UseStatus().ValidPanels() {
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
				JC.UseWorker().ResumeAll()
				JA.UseStatus().Resume()
			}

			if !JA.UseStatus().IsReady() {
				JC.Logln("Refused to fetch data as app is not ready yet")
				return
			}

			if !JA.UseStatus().HasError() && JC.IsMobile {
				// Force Refresh
				JT.UseExchangeCache().SoftReset()
				JC.UseWorker().Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.UseTickerCache().SoftReset()
				JC.UseWorker().Call("update_tickers", JC.CallImmediate)
			}
		})
		lc.SetOnExitedForeground(func() {
			JC.Logln("App exited foreground")

			if JC.IsMobile {
				JC.Logln("Battery Saver: Pausing apps")
				JA.UseStatus().Pause()
				JC.UseWorker().PauseAll()
			}

			if !JA.UseStatus().IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved && JC.IsMobile {
				JA.UseSnapshot().Save()
				snapshotSaved = true
			}
		})
		lc.SetOnStopped(func() {
			JC.Logln("App stopped")

			if !JA.UseStatus().IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved {
				JA.UseSnapshot().Save()
				snapshotSaved = true
			}
		})
	}
}

func RegisterDispatcher() {
	d := core.UseDispatcher()
	d.Init()
	d.SetBufferSize(10000)

	if JC.IsMobile {
		d.SetMaxConcurrent(1)
		d.SetDelayBetween(48 * time.Millisecond)

	} else {
		d.SetMaxConcurrent(10)
		d.SetDelayBetween(1 * time.Millisecond)
	}

	d.Start()
}

func RegisterCache() {
	// Prepopulating character sizes
	sizes := []float32{
		JC.ThemeSize(JC.SizePanelTitle),
		JC.ThemeSize(JC.SizePanelSubTitle),
		JC.ThemeSize(JC.SizePanelBottomText),
		JC.ThemeSize(JC.SizePanelContent),
		JC.ThemeSize(JC.SizePanelTitleSmall),
		JC.ThemeSize(JC.SizePanelSubTitleSmall),
		JC.ThemeSize(JC.SizePanelBottomTextSmall),
		JC.ThemeSize(JC.SizePanelContentSmall),
		JC.ThemeSize(JC.SizeTickerTitle),
		JC.ThemeSize(JC.SizeTickerContent),
		JC.ThemeSize(JC.SizeNotificationText),
		JC.ThemeSize(JC.SizeCompletionText),
	}

	for _, size := range sizes {
		// Normal
		styleBits := 0
		key := int(size)*10 + styleBits
		if !JC.UseCharWidthCache().Has(key) {
			JC.UseCharWidthCache().Add(key, fyne.MeasureText("a", size, fyne.TextStyle{}).Width)
		}

		// Bold
		styleBits = 1
		key = int(size)*10 + styleBits
		if !JC.UseCharWidthCache().Has(key) {
			JC.UseCharWidthCache().Add(key, fyne.MeasureText("a", size, fyne.TextStyle{Bold: true}).Width)
		}

		// Italic
		styleBits = 2
		key = int(size)*10 + styleBits
		if !JC.UseCharWidthCache().Has(key) {
			JC.UseCharWidthCache().Add(key, fyne.MeasureText("a", size, fyne.TextStyle{Italic: true}).Width)
		}

		// Monospace
		styleBits = 4
		key = int(size)*10 + styleBits
		if !JC.UseCharWidthCache().Has(key) {
			JC.UseCharWidthCache().Add(key, fyne.MeasureText("a", size, fyne.TextStyle{Monospace: true}).Width)
		}
	}
}
