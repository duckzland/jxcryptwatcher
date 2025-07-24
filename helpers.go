package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
)

func UpdateDisplay() {

	list := JT.BP.Get()
	for i := range list {
		// Always get linked data! do not use the copied
		pkt := JT.BP.GetDataByIndex(i)
		pk := pkt.Get()
		pkt.Update(pk)
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
	fyne.Do(func() {
		JC.Grid.Refresh()
		UpdateDisplay()
	})

	if JT.SavePanels() {
		if UpdateRates() {
			fyne.Do(func() {
				UpdateDisplay()
			})
		}
	}
}

func OpenNewPanelForm() {
	d := JP.NewPanelForm(
		"new",
		func() {
			SavePanelForm()
		},
		func(npdt *JT.PanelDataType) {
			JC.Grid.Add(JP.NewPanelNormal(
				npdt,

				// Open the panel edit callback
				func(dynpk string) {
					JP.NewPanelForm(
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
	d := JP.NewPanelForm(pk, func() {
		SavePanelForm()
	}, nil)

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
	return JP.NewPanelNormal(
		pkt,
		func(dynpk string) {
			OpenPanelEditForm(dynpk)
		},
		func(dynpi int) {
			RemovePanelByIndex(dynpi)
		},
	)
}
