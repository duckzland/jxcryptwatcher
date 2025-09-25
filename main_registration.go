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
	JC.UseWorker().RegisterListener(
		"update_display", 200,
		func() {
			if UpdateDisplay() {
				JA.UseLayout().RegisterDisplayUpdate(time.Now())
			}
		},
		func() bool {
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
			if !JT.UseExchangeCache().GetTimestamp().After(JA.UseLayout().GetDisplayUpdate()) {
				JC.Notify("Unable to refresh display: Data is older than display timestamp")
				return false
			}
			if !JA.UseStatus().ValidPanels() {
				JC.Logln("Unable to refresh display: No valid panels configured")
				return false
			}
			return true
		},
	)

	JC.UseWorker().Register(
		"update_rates", 1000,
		func() {
			UpdateRates()
		},
		func() int64 {
			return max(JT.UseConfig().Delay*1000, 30000)
		},
		func() bool {
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
		},
	)

	JC.UseWorker().Register(
		"update_tickers", 5000,
		func() {
			UpdateTickers()
		},
		func() int64 {
			return max(JT.UseConfig().Delay*1000, 30000)
		},
		func() bool {
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
		},
	)

	JC.UseWorker().RegisterBuffered(
		"notification", 1000, 2000,
		func() int64 {
			return 1000
		},
		func(messages []string) bool {
			if len(messages) == 0 {
				return false
			}
			latest := messages[len(messages)-1]
			fyne.Do(func() {
				JW.UseNotification().UpdateText(latest)
			})
			ScheduledNotificationReset()
			return true
		},
		func() bool {
			if !JA.UseStatus().IsReady() {
				JC.Logln("Unable to do notification: app is not ready yet")
				return false
			}
			if JA.UseStatus().IsPaused() {
				JC.Logln("Unable to do notification: app is paused")
				return false
			}
			return true
		},
	)
}

func RegisterFetchers() {
	var delay int64 = 100

	JC.UseFetcher().Register(
		"cryptos_map", 0,
		JC.NewGenericFetcher(
			func(ctx context.Context) (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(
					JT.UseCryptosLoader().GetCryptos(),
					nil,
				), nil
			},
		),
		func(result JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingCryptos()

			status := DetectHTTPResponse(result.Code())
			ProcessFetchingCryptosComplete(status)
		},
		func() bool {
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
		},
	)

	JC.UseFetcher().Register(
		"cmc100", delay,
		JC.NewGenericFetcher(
			func(ctx context.Context) (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewCMC100Fetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
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
		},
	)

	JC.UseFetcher().Register(
		"market_cap", delay,
		JC.NewGenericFetcher(
			func(ctx context.Context) (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewMarketCapFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
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
		},
	)

	JC.UseFetcher().Register(
		"altcoin_index", delay,
		JC.NewGenericFetcher(
			func(ctx context.Context) (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewAltSeasonFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
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
		},
	)

	JC.UseFetcher().Register(
		"feargreed", delay,
		JC.NewGenericFetcher(
			func(ctx context.Context) (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewFearGreedFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
			if !JA.UseStatus().IsReady() {
				JC.Logln("Unable to fetch feargreed: app is not ready yet")
				return false
			}
			if JA.UseStatus().IsPaused() {
				JC.Logln("Unable to fetch feargreed: app is paused")
				return false
			}
			if !JT.UseConfig().CanDoFearGreed() {
				JC.Logln("Unable to fetch feargreed: Invalid config")
				return false
			}
			return true
		},
	)

	JC.UseFetcher().Register(
		"rates", delay,
		JC.NewDynamicPayloadFetcher(
			func(ctx context.Context, payload any) (JC.FetchResultInterface, error) {
				rk, ok := payload.(string)
				if !ok {
					return JC.NewFetchResult(JC.NETWORKING_BAD_PAYLOAD, nil), fmt.Errorf("invalid rk")
				}

				return JC.NewFetchResult(JT.NewExchangeResults().GetRate(rk), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
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
		},
	)
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
	JC.PrintPerfStats("Creating Dispatcher Buffer", time.Now())
	d := JC.UseDispatcher()
	d.SetBufferSize(1000000)
	d.SetDelayBetween(100 * time.Millisecond)

	if JC.IsMobile {
		d.SetMaxConcurrent(JC.MaximumThreads(4))

	} else {
		d.SetMaxConcurrent(JC.MaximumThreads(6))
	}

	d.Start()
}

func RegisterCache() {
	// Prepopulating character sizes
	sizes := []float32{
		JC.UseTheme().Size(JC.SizePanelTitle),
		JC.UseTheme().Size(JC.SizePanelSubTitle),
		JC.UseTheme().Size(JC.SizePanelBottomText),
		JC.UseTheme().Size(JC.SizePanelContent),
		JC.UseTheme().Size(JC.SizePanelTitleSmall),
		JC.UseTheme().Size(JC.SizePanelSubTitleSmall),
		JC.UseTheme().Size(JC.SizePanelBottomTextSmall),
		JC.UseTheme().Size(JC.SizePanelContentSmall),
		JC.UseTheme().Size(JC.SizeTickerTitle),
		JC.UseTheme().Size(JC.SizeTickerContent),
		JC.UseTheme().Size(JC.SizeNotificationText),
		JC.UseTheme().Size(JC.SizeCompletionText),
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
