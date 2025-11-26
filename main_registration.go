package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	JX "jxwatcher/animations"
	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JS "jxwatcher/tickers"
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

func registerBoot() {

	JC.InitLogger()

	JC.Logln("App is booting...")

	if !JC.IsMobile {
		configDir := JC.GetUserDirectory()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			JC.Logf("Error creating directory: %v", err)
		}
	}

	JC.Window.SetOnClosed(func() {
		JC.Logln("Window was closed")
		JC.App.Quit()
		JC.ShutdownSignal <- syscall.SIGTERM
	})
}

func registerTheme() {
	JC.RegisterThemeManager().Init()
	// Comment this out for now, as we dont have real settings to force DarkTheme
	// JC.UseTheme().SetVariant(JC.App.Settings().ThemeVariant())
	JC.App.Settings().SetTheme(JC.UseTheme())
}

func registerAppIcon() {
	icon := fyne.NewStaticResource("jxwatcher.png", appIconData)
	JC.Window.SetIcon(icon)
}

func registerFonts() {
	JC.UseTheme().SetFonts(fyne.TextStyle{Bold: false}, fyne.NewStaticResource("Roboto-Regular.ttf", regularFont))
	JC.UseTheme().SetFonts(fyne.TextStyle{Bold: true}, fyne.NewStaticResource("Roboto-Bold.ttf", boldFont))
}

func registerUtility() {
	JC.RegisterDebouncer().Init()
	JA.RegisterSnapshotManager().Init()
	JA.RegisterStatusManager().Init()
}

func registerActions() {

	JA.RegisterActionManager().Init()

	// Refresh ticker data
	JA.UseAction().Add(JW.NewActionButton(JC.ACT_CRYPTO_REFRESH_MAP, JC.STRING_EMPTY, theme.ViewRestoreIcon(), "Refresh cryptos data", "disabled",
		func(btn JW.ActionButton) {
			JC.UseFetcher().Call(JC.ACT_CRYPTO_GET_MAP, nil)
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
	JA.UseAction().Add(JW.NewActionButton(JC.ACT_EXCHANGE_REFRESH_RATES, JC.STRING_EMPTY, theme.ViewRefreshIcon(), "Update rates from exchange", "disabled",
		func(btn JW.ActionButton) {
			// Open the network status temporarily
			JA.UseStatus().SetNetworkStatus(true)

			// Force update
			JT.UseExchangeCache().SoftReset()
			JC.UseWorker().Flush(JC.ACT_EXCHANGE_UPDATE_RATES)
			JC.UseWorker().Call(JC.ACT_EXCHANGE_UPDATE_RATES, JC.CallDebounced)

			// Force update
			JT.UseTickerCache().SoftReset()
			JC.UseWorker().Flush(JC.ACT_TICKER_UPDATE)
			JC.UseWorker().Call(JC.ACT_TICKER_UPDATE, JC.CallDebounced)

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
	JA.UseAction().Add(JW.NewActionButton(JC.ACT_OPEN_SETTINGS, JC.STRING_EMPTY, theme.SettingsIcon(), "Open settings", "disabled",
		func(btn JW.ActionButton) {
			openSettingForm()
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

			if !JA.UseStatus().IsGoodNetworkStatus() {
				btn.Error()
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

			if JA.UseStatus().IsFetchingCryptos() {
				btn.Disable()
				return
			}

			if JA.UseStatus().IsFetchingRates() {
				btn.Disable()
				return
			}

			btn.Enable()
		}))

	// Panel drag toggle
	JA.UseAction().Add(JW.NewActionButton(JC.ACT_PANEL_DRAG, JC.STRING_EMPTY, theme.ContentPasteIcon(), "Enable Reordering", "disabled",
		func(btn JW.ActionButton) {
			toggleDraggable()
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
	JA.UseAction().Add(JW.NewActionButton(JC.ACT_PANEL_ADD, JC.STRING_EMPTY, theme.ContentAddIcon(), "Add new panel", "disabled",
		func(btn JW.ActionButton) {
			openNewPanelForm()
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

func registerWorkers() {

	JC.RegisterWorkerManager().Init()

	JC.UseWorker().Register(
		JC.ACT_PANEL_UPDATE, 1,
		func() int64 {
			return 200
		},
		nil,
		func(any) bool {
			if updateDisplay() {
				JA.UseLayout().RegisterDisplayUpdate(time.Now())
				return true
			}
			return false
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
				JC.Logln("Unable to refresh display: no cached data")
				return false
			}
			if !JT.UseExchangeCache().GetTimestamp().After(JA.UseLayout().GetDisplayUpdate()) {
				JC.Logln("Unable to refresh display: Data is older than display timestamp")
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
		JC.ACT_EXCHANGE_UPDATE_RATES, 1,
		nil,
		func() int64 {
			return max(JT.UseConfig().Delay*1000, 30000)
		},
		func(any) bool {
			return updateRates()
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
				JC.Logln("Unable to refresh rates: invalid configuration")
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
		JC.ACT_TICKER_UPDATE, 1,
		nil,
		func() int64 {
			return max(JT.UseConfig().Delay*1000, 30000)
		},
		func(any) bool {
			return updateTickers()
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

	JC.UseWorker().Register(
		JC.ACT_NOTIFICATION_PUSH, 50,
		func() int64 {
			return 1000
		},
		nil,
		func(payload any) bool {
			latest, _ := payload.(string)
			fyne.Do(func() {
				JW.UseNotification().UpdateText(latest)
			})

			scheduledNotificationReset()
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

func registerFetchers() {
	var delay int64 = 100

	JC.RegisterFetcherManager().Init()

	JC.UseFetcher().Register(
		JC.ACT_CRYPTO_GET_MAP, 0,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(
					JT.UseCryptosLoader().GetCryptos(),
					nil,
				), nil
			},
		),
		func(result JC.FetchResultInterface) {
			defer JA.UseStatus().EndFetchingCryptos()

			status := detectHTTPResponse(result.Code())
			processFetchingCryptosComplete(status)
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
				JC.Notify(JC.NotifyInvalidConfigurationUnableToResetCryptos)
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
		JT.TickerTypeCMC100, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
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
		JT.TickerTypeMarketCap, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
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
		JT.TickerTypeAltcoinIndex, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
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
		JT.TickerTypeFearGreed, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
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
		JT.TickerTypeRSI, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewRSIFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
			if !JA.UseStatus().IsReady() {
				JC.Logln("Unable to fetch rsi: app is not ready yet")
				return false
			}
			if JA.UseStatus().IsPaused() {
				JC.Logln("Unable to fetch rsi: app is paused")
				return false
			}
			if !JT.UseConfig().CanDoFearGreed() {
				JC.Logln("Unable to fetch rsi: Invalid config")
				return false
			}
			return true
		},
	)

	JC.UseFetcher().Register(
		JT.TickerTypeETF, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewETFFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
			if !JA.UseStatus().IsReady() {
				JC.Logln("Unable to fetch etf: app is not ready yet")
				return false
			}
			if JA.UseStatus().IsPaused() {
				JC.Logln("Unable to fetch etf: app is paused")
				return false
			}
			if !JT.UseConfig().CanDoETF() {
				JC.Logln("Unable to fetch etf: Invalid config")
				return false
			}
			return true
		},
	)

	JC.UseFetcher().Register(
		JT.TickerTypeDominance, delay,
		JC.NewGenericFetcher(
			func() (JC.FetchResultInterface, error) {
				return JC.NewFetchResult(JT.NewDominanceFetcher().GetRate(), nil), nil
			},
		),
		func(fr JC.FetchResultInterface) {
			// Results is processed at GetRate()
		},
		func() bool {
			if !JA.UseStatus().IsReady() {
				JC.Logln("Unable to fetch dominance: app is not ready yet")
				return false
			}
			if JA.UseStatus().IsPaused() {
				JC.Logln("Unable to fetch dominance: app is paused")
				return false
			}
			if !JT.UseConfig().CanDoDominance() {
				JC.Logln("Unable to fetch dominance: Invalid config")
				return false
			}
			return true
		},
	)

	JC.UseFetcher().Register(
		JC.ACT_EXCHANGE_GET_RATES, delay,
		JC.NewDynamicPayloadFetcher(
			func(payload any) (JC.FetchResultInterface, error) {
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

func registerLifecycle() {

	var isAppStarted bool = false

	// Hook into lifecycle events
	if lc := JC.App.Lifecycle(); lc != nil {

		lc.SetOnStarted(func() {

			if isAppStarted {
				JC.Logln("App is already started, refused to restart it again")
				return
			}

			JC.Logln("App started")
			JC.Notify(JC.NotifyApplicationIsStarting)

			JA.RegisterLayoutManager().Init()

			JC.Window.SetContent(JA.NewAppLayout())

			// Prevent locking when initialized at first install
			JC.UseDebouncer().Call("initializing", 1*time.Millisecond, func() {

				if !JT.ConfigInit() {
				}

				if JA.UseSnapshot().LoadCryptos() == JC.NO_SNAPSHOT {
					JT.CryptosLoaderInit()
				}

				if JA.UseSnapshot().LoadExchangeData() == JC.NO_SNAPSHOT {
					JT.UseExchangeCache().Reset()
				}

				if JA.UseSnapshot().LoadTickerData() == JC.NO_SNAPSHOT {
					JT.UseTickerCache().Reset()
				}

				if JA.UseSnapshot().LoadPanels() == JC.NO_SNAPSHOT {
					JT.PanelsInit()
				}

				if JA.UseSnapshot().LoadTickers() == JC.NO_SNAPSHOT {
					JT.TickersInit()
				}

				JT.UseConfig().PostInit()

				fyne.Do(func() {

					JS.RegisterTickerGrid()
					JP.RegisterPanelGrid(createPanel)

					JA.UseStatus().InitData()
					JA.UseLayout().RegisterContent(JP.UsePanelGrid())
					JP.UsePanelGrid().Refresh()
					JA.UseLayout().UpdateState()

					JC.Logln("App is ready: ", JA.UseStatus().IsReady())

					if !JA.UseStatus().HasError() {

						// Force Refresh
						JT.UseExchangeCache().SoftReset()
						JC.UseWorker().Call(JC.ACT_EXCHANGE_UPDATE_RATES, JC.CallImmediate)

						// Force Refresh
						JT.UseTickerCache().SoftReset()
						JC.UseWorker().Call(JC.ACT_TICKER_UPDATE, JC.CallImmediate)
					}
				})

				isAppStarted = true
			})
		})

		lc.SetOnEnteredForeground(func() {
			if !isAppStarted {
				JC.Logln("App is not started yet, refuse to init app entered foreground")
				return
			}

			JC.Logln("App entered foreground")

			JA.UseSnapshot().Reset()

			if JC.IsMobile {
				JC.Logln("Battery Saver: Continuing apps")
				JC.UseWorker().Resume()
				JA.UseStatus().Resume()
				JC.UseDispatcher().Resume()
			}

			if !JA.UseStatus().IsReady() {
				JC.Logln("Refused to fetch data as app is not ready yet")
				return
			}

			if !JA.UseStatus().HasError() && JC.IsMobile {
				// Force Refresh
				JT.UseExchangeCache().SoftReset()
				JC.UseWorker().Call(JC.ACT_EXCHANGE_UPDATE_RATES, JC.CallImmediate)

				// Force Refresh
				JT.UseTickerCache().SoftReset()
				JC.UseWorker().Call(JC.ACT_TICKER_UPDATE, JC.CallImmediate)
			}
		})

		lc.SetOnExitedForeground(func() {
			if !isAppStarted {
				JC.Logln("App is not started yet, refuse to init app exited foreground")
				return
			}

			JC.Logln("App exited foreground")

			if JC.IsMobile {
				JC.Logln("Battery Saver: Pausing apps")
				JA.UseStatus().Pause()
				JC.UseWorker().Pause()
				JC.UseDispatcher().Pause()
			}

			if !JA.UseStatus().IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !JA.UseSnapshot().IsSnapshotted() && JC.IsMobile {
				JA.UseSnapshot().Save()
			}
		})

		lc.SetOnStopped(func() {
			JC.Logln("App stopped")
		})
	}
}

func registerDispatcher() {
	JC.PrintPerfStats("Creating Dispatcher Buffer", time.Now())

	JC.RegisterDispatcher().Init()
	JX.RegisterAnimationDispatcher().Init()

	ad := JX.UseAnimationDispatcher()
	ad.SetBufferSize(100)

	if JC.IsMobile {
		ad.SetDelayBetween(200 * time.Millisecond)
		ad.SetMaxConcurrent(2)

	} else {
		ad.SetDelayBetween(50 * time.Millisecond)
		ad.SetMaxConcurrent(JC.MaximumThreads(6))
	}

	ad.Start()

	d := JC.UseDispatcher()
	d.SetBufferSize(100)

	if JC.IsMobile {
		d.SetDelayBetween(10 * time.Millisecond)
		d.SetMaxConcurrent(2)

	} else {
		d.SetDelayBetween(5 * time.Millisecond)
		d.SetMaxConcurrent(JC.MaximumThreads(6))
	}

	d.Start()
}

func registerCache() {

	JT.RegisterExchangeCache().Init()

	JT.RegisterTickerCache().Init()

	JC.RegisterCharWidthCache().Init()

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
			JC.UseCharWidthCache().Add(key, JC.MeasureText("a", size, fyne.TextStyle{}))
		}

		// Bold
		styleBits = 1
		key = int(size)*10 + styleBits
		if !JC.UseCharWidthCache().Has(key) {
			JC.UseCharWidthCache().Add(key, JC.MeasureText("a", size, fyne.TextStyle{Bold: true}))
		}

		// Italic
		// styleBits = 2
		// key = int(size)*10 + styleBits
		// if !JC.UseCharWidthCache().Has(key) {
		// 	JC.UseCharWidthCache().Add(key, JC.MeasureText("a", size, fyne.TextStyle{Italic: true}))
		// }

		// Monospace
		// styleBits = 4
		// key = int(size)*10 + styleBits
		// if !JC.UseCharWidthCache().Has(key) {
		// 	JC.UseCharWidthCache().Add(key, JC.MeasureText("a", size, fyne.TextStyle{Monospace: true}))
		// }
	}
}
