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

		// Give pause to prevent race condition
		time.Sleep(1 * time.Millisecond)
	}

	// log.Print("Display Refreshed")
}

func UpdateRates() bool {

	// Clear cached rates
	JT.ExchangeCache.Reset()

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
		SavePanelForm,
		func(npdt *JT.PanelDataType) {
			JC.Grid.Add(JP.NewPanelDisplay(
				npdt,

				// Open the panel edit callback
				func(dynpk string) {
					JP.NewPanelForm(
						dynpk,
						// Save panel form callback
						SavePanelForm,
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
	return JP.NewPanelDisplay(
		pkt,
		func(dynpk string) {
			OpenPanelEditForm(dynpk)
		},
		func(dynpi int) {
			RemovePanelByIndex(dynpi)
		},
	)
}
