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

	if len(jb) == 0 {
		return false
	}

	JC.Notify("Start retrieving rates...")

	// Fetching with delay
	for _, rk := range jb {
		ex.GetRate(rk)

		// Give pause to prevent too many connection open at once
		time.Sleep(200 * time.Millisecond)
	}

	JC.Notify("New rates retrieved")

	log.Printf("Exchange Rate updated: %v/%v", len(jb), len(list))

	return true
}

func RefreshRates() bool {

	if !JT.Config.IsValid() {
		log.Println("Invalid configuration, cannot refresh rates")
		JC.Notify("Invalid configuration, cannot refresh rates")
		return false
	}

	// Clear cached rates
	JT.ExchangeCache.Reset()

	if UpdateRates() {
		fyne.Do(UpdateDisplay)
	}

	return true
}

func RemovePanel(uuid string) {

	for _, obj := range JC.Grid.Objects {
		if panel, ok := obj.(*JW.DoubleClickContainer); ok {
			if panel.GetTag() == uuid {

				log.Printf("Removing panel %s", uuid)

				JC.Grid.Remove(obj)

				fyne.Do(JC.Grid.Refresh)

				if JT.BP.Remove(uuid) {
					if JT.SavePanels() {
						JC.Notify("Panel removed")
					}
				}
			}
		}
	}
}

func SavePanelForm() {

	JC.Notify("Saving panel configuration...")

	fyne.Do(func() {
		JC.Grid.Refresh()
		JC.RequestDisplayUpdate()
	})

	go func() {
		if JT.SavePanels() {
			if UpdateRates() {
				JC.RequestDisplayUpdate()
				JC.Notify("Panel configuration saved")
			}

		} else {
			JC.Notify("Failed to save panel")
		}
	}()
}

func OpenNewPanelForm() {
	d := JP.NewPanelForm(
		"new",
		"",
		SavePanelForm,
		func(npdt *JT.PanelDataType) {
			fyne.Do(func() {
				JC.Grid.Add(CreatePanel(npdt))
				JC.Notify("New Panel created")
			})
		},
	)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenPanelEditForm(pk string, uuid string) {
	d := JP.NewPanelForm(pk, uuid, SavePanelForm, nil)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenSettingForm() {
	d := JA.NewSettingsForm(func() {
		JC.Notify("Saving configuration...")

		if JT.Config.SaveFile() != nil {
			JC.Notify("Configuration data saved...")
		} else {
			JC.Notify("Failed to save configuration")
		}
	})

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func CreatePanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JP.NewPanelDisplay(pkt, OpenPanelEditForm, RemovePanel)
}

func ResetCryptosMap() {
	if !JT.Config.IsValid() {
		log.Println("Invalid configuration, cannot reset cryptos map")
		JC.Notify("Invalid configuration, cannot reset cryptos map")
		return
	}

	Cryptos := JT.CryptosType{}
	JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
	JT.BP.Maps.ClearMapCache()

	JC.Notify("Cryptos map regenerated")

	if JT.BP.RefreshData() {
		fyne.Do(func() {
			JC.Grid.Refresh()
			JC.RequestDisplayUpdate()
		})
	}
}

func StartWorkers() {
	go func() {
		for range JC.UpdateDisplayChan {
			if !JT.Config.IsValid() {
				continue
			}

			fyne.Do(UpdateDisplay)
		}
	}()
	go func() {
		for range JC.UpdateRatesChan {
			if !JT.Config.IsValid() {
				continue
			}

			RefreshRates()
		}
	}()
}

func StartUpdateDisplayWorker() {
	go func() {
		for {
			JC.RequestDisplayUpdate()
			time.Sleep(3 * time.Second)
		}
	}()
}

func StartUpdateRatesWorker() {
	go func() {
		for {
			JC.RequestRateUpdate()
			time.Sleep(time.Duration(JT.Config.Delay) * time.Second)
		}
	}()
}
