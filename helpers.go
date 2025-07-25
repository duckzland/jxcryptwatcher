package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func UpdateDisplay() {

	list := JT.BP.Get()
	for i := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetDataByIndex(i)
		pk := pkt.Get()
		pkt.Update(pk)

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	// log.Print("Display Refreshed")
}

func UpdateRates() bool {
	ex := JT.ExchangeResults{}
	jb := make(map[string]string)
	list := JT.BP.Get()

	// Prune data first, remove duplicate calls, merge into single call wheneveer possible
	for i := range list {
		pk := JT.BP.GetDataByIndex(i)
		pkt := pk.UsePanelKey()
		sid := pkt.GetSourceCoinString()
		tid := pkt.GetTargetCoinString()

		_, exists := jb[sid]
		if !exists {
			jb[sid] = sid + "|" + tid
		} else {
			jb[sid] += "," + tid

		}
	}

	// Fetching with delay
	for _, rk := range jb {
		ex.GetRate(rk)

		// Give pause to prevent too many connection open at once
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Exchange Rate updated: %v/%v", len(jb), len(list))

	return true
}

func RefreshRates() {
	// Clear cached rates
	JT.ExchangeCache.Reset()

	if UpdateRates() {
		fyne.Do(UpdateDisplay)
	}
}

func RemovePanelByIndex(di int) {
	if JT.RemovePanel(di) {
		if di >= 0 && di < len(JC.Grid.Objects) {
			JC.Grid.Objects = append(JC.Grid.Objects[:di], JC.Grid.Objects[di+1:]...)
		}
		fyne.Do(func() {
			JC.Grid.Refresh()
		})
		JT.SavePanels()
	}
}

func SavePanelForm() {

	fyne.Do(func() {
		JC.Grid.Refresh()
		JC.UpdateDisplayChan <- struct{}{}
	})

	if JT.SavePanels() {
		if UpdateRates() {
			JC.UpdateDisplayChan <- struct{}{}
		}
	}
}

func OpenNewPanelForm() {
	d := JP.NewPanelForm(
		"new",
		SavePanelForm,
		func(npdt *JT.PanelDataType) {
			JC.Grid.Add(CreatePanel(npdt))
		},
	)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenPanelEditForm(pk string) {
	d := JP.NewPanelForm(pk, SavePanelForm, nil)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenSettingForm() {
	d := JA.NewSettingsForm(func() {
		JT.Config.SaveFile()
	})

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanelByIndex)
}

func ResetCryptosMap() {
	Cryptos := JT.CryptosType{}
	JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
	JT.BP.Maps.ClearMapCache()
	JC.UpdateDisplayChan <- struct{}{}
}

func StartWorkers() {
	go func() {
		for range JC.UpdateDisplayChan {
			fyne.Do(UpdateDisplay)
		}
	}()
	go func() {
		for range JC.UpdateRatesChan {
			JW.DoActionWithNotification(
				"Fetching exchange rate...",
				"Fetching rates from exchange...",
				JC.NotificationBox,
				RefreshRates,
			)
		}
	}()
}

func StartUpdateDisplayWorker() {
	go func() {
		for {
			JC.UpdateDisplayChan <- struct{}{}
			time.Sleep(3 * time.Second)
		}
	}()
}

func StartUpdateRatesWorker() {
	go func() {
		for {
			JC.UpdateRatesChan <- struct{}{}
			time.Sleep(time.Duration(JT.Config.Delay) * time.Second)
		}
	}()
}
