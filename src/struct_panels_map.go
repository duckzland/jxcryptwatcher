package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

type PanelsMap struct {
	data []PanelDataType
	maps CryptosMapType
}

func (pc *PanelsMap) Init() {
	pc.data = []PanelDataType{}
}

func (pc *PanelsMap) Set(data []PanelDataType) {
	pc.data = data
}

func (pc *PanelsMap) SetMaps(maps CryptosMapType) {
	pc.maps = maps
}

func (pc *PanelsMap) Remove(index int) bool {
	values := pc.data
	if index < 0 || index >= len(values) {
		return false
	}

	pc.data = append(values[:index], values[index+1:]...)

	return true
}

func (pc *PanelsMap) RemoveByKey(pk string) bool {
	return pc.Remove(pc.GetIndex(pk))
}

func (pc *PanelsMap) Append(pk string) *PanelDataType {

	if pc.data == nil {
		pc.data = []PanelDataType{}
	}

	pn := PanelDataType{
		data: binding.NewString(),
		// oldKey: pk,
		parent: pc,
		index:  -1,
	}

	pn.Update(pk)
	pc.data = append(pc.data, pn)

	return &pn
}

func (pc *PanelsMap) Update(pk string, index int) *PanelDataType {

	if index < 0 || index >= len(pc.data) {
		return nil
	}

	pdt := &pc.data[index]

	pdt.Update(pk)

	return pdt
}

func (pc *PanelsMap) Get() []PanelDataType {
	return pc.data
}

func (pc *PanelsMap) GenerateKey(source, target, value, decimals string, rate float32) string {
	return fmt.Sprintf("%s-%s-%s-%s|%f", source, target, value, decimals, rate)
}

func (pc *PanelsMap) GenerateKeyFromPanel(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%s-%d|%f", panel.Source, panel.Target, dynamicFormatFloatToString(panel.Value), panel.Decimals, rate)
}

func (pc *PanelsMap) GetIndex(pk string) int {
	list := pc.data
	for i, pdt := range list {
		if pdt.IsEqualContentString(pk) {
			return i
		}
	}

	return -1
}

func (pc *PanelsMap) GetDataByIndex(index int) *PanelDataType {
	// If index is out of bounds, return nil
	if index < 0 || index >= len(pc.data) {
		return nil
	}

	return &pc.data[index]
}

func (pc *PanelsMap) GetSourceCoin(pk string) int64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		source, err := strconv.ParseInt(pkv[0], 10, 64)
		if err == nil {
			return source
		}
	}
	return 0
}

func (pc *PanelsMap) GetTargetCoin(pk string) int64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		target, err := strconv.ParseInt(pkv[1], 10, 64)
		if err == nil {
			return target
		}
	}

	return 0
}

func (pc *PanelsMap) ValidateKey(pk string) bool {
	pkv := strings.Split(pk, "|")
	if len(pkv) != 2 {
		return false
	}

	pkt := strings.Split(pkv[0], "-")
	if len(pkt) != 4 {
		return false
	}

	return true
}

func (pc *PanelsMap) ValidatePanel(pk string) bool {
	if !pc.ValidateKey(pk) {
		return false
	}

	sid := pc.GetSourceCoin(pk)
	tid := pc.GetTargetCoin(pk)

	// @todo when cryptos got method use that!
	if !pc.maps.ValidateId(sid) {
		return false
	}

	// @todo when cryptos got method use that!
	if !pc.maps.ValidateId(tid) {
		return false
	}

	return true
}

func (pc *PanelsMap) ValidateId(id int64) bool {
	return pc.maps.ValidateId(id)
}

func (pc *PanelsMap) GetOptions() []string {
	return pc.maps.GetOptions()
}

func (pc *PanelsMap) GetDisplayById(id string) string {
	return pc.maps.GetDisplayById(id)
}

func (pc *PanelsMap) GetIdByDisplay(id string) string {
	return pc.maps.GetIdByDisplay(id)
}

func (pc *PanelsMap) GetSymbolById(id string) string {
	return pc.maps.GetSymbolById(id)
}
