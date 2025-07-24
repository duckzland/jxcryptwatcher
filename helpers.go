package main

import (
	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
	"log"
	"time"

	"fyne.io/fyne/v2"
)

func UpdateDisplay() {

	list := JT.BP.Get()
	for i := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetDataByIndex(i)
		pk := pkt.Get()

		if JT.BP.ValidatePanel(pk) {
			if pkt.Update(pk) {
				if pkt.Index == -1 {
					npk := pkt.Get()
					// This panel hasnt been generated yet, create the markup!
					if JT.BP.ValidatePanel(npk) {
						JC.Grid.Objects[i] = CreateNormalPanel(pkt)
					} else {
						JC.Grid.Objects[i] = CreateInvalidPanel(pk)
					}
				}
			}
		} else {
			if pkt.Index == -1 {
				JC.Grid.Objects[i] = CreateInvalidPanel(pk)
			}
		}

		pkt.Index = i
	}

	// log.Print("Display Refreshed")
}

func UpdateRates() bool {

	// Clear cached rates
	JT.ExchangeCache.Reset()

	ex := JT.ExchangeDataType{}
	jb := make(map[string]string)
	list := JT.BP.Get()

	// Prune data first, remove duplicate calls
	for _, pk := range list {
		ck := JT.ExchangeCache.CreateKeyFromString(
			pk.UsePanelKey().GetSourceCoinString(),
			pk.UsePanelKey().GetTargetCoinString(),
		)

		_, exists := jb[ck]
		if !exists {
			jb[ck] = pk.Get()
		}
	}

	// Fetching with delay
	for _, pk := range jb {
		ex.GetRate(pk)

		// Give pause to prevent too many connection open at once
		time.Sleep(200 * time.Millisecond)
	}

	log.Print("Exchange Rate updated")

	return true
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
	if JT.SavePanels() {
		fyne.Do(func() {
			JC.Grid.Refresh()
		})
		if UpdateRates() {
			fyne.Do(func() {
				UpdateDisplay()
			})
		}
	}
}

func OpenNewPanelForm() {
	d := JW.NewPanelForm(
		"new",
		func() {
			SavePanelForm()
		},
		func(npdt *JT.PanelDataType) {
			JC.Grid.Add(JW.NewPanel(
				npdt,

				// Open the panel edit callback
				func(dynpk string) {
					JW.NewPanelForm(
						dynpk,
						// Save panel form callback
						func() {
							SavePanelForm()
						},
						// No new panel callback
						nil,
					)
				},

				// Delete Callback
				func(dynpi int) {
					RemovePanelByIndex(dynpi)
				},
			))
		},
	)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenPanelEditForm(pk string) {
	d := JW.NewPanelForm(pk, func() {
		SavePanelForm()
	}, nil)

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
}

func OpenSettingForm() {
	d := JW.NewSettingsForm(func() {
		JT.Config.SaveFile()
	})

	d.Show()
	d.Resize(fyne.NewSize(400, 300))
	log.Printf("Triggered", d)
}

func CreateInvalidPanel(pk string) fyne.CanvasObject {
	return JW.NewInvalidPanel(
		pk,
		func(dpk string) {
			OpenPanelEditForm(dpk)
		},
		func(di int) {
			RemovePanelByIndex(di)
		},
	)
}

func CreateNormalPanel(pkt *JT.PanelDataType) fyne.CanvasObject {
	return JW.NewPanel(
		pkt,
		func(dynpk string) {
			OpenPanelEditForm(dynpk)
		},
		func(dynpi int) {
			RemovePanelByIndex(dynpi)
		},
	)
}
