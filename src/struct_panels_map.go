package main

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

type PanelsMapType struct {
	data []PanelDataType
	maps CryptosMapType
}

func (pc *PanelsMapType) Init() {
	pc.data = []PanelDataType{}
}

func (pc *PanelsMapType) Set(data []PanelDataType) {
	pc.data = data
}

func (pc *PanelsMapType) SetMaps(maps CryptosMapType) {
	pc.maps = maps
}

func (pc *PanelsMapType) Remove(index int) bool {
	values := pc.data
	if index < 0 || index >= len(values) {
		return false
	}

	pc.data = append(values[:index], values[index+1:]...)

	return true
}

func (pc *PanelsMapType) RemoveByKey(pk string) bool {
	return pc.Remove(pc.GetIndex(pk))
}

func (pc *PanelsMapType) Append(pk string) *PanelDataType {

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

func (pc *PanelsMapType) Update(pk string, index int) *PanelDataType {

	if index < 0 || index >= len(pc.data) {
		return nil
	}

	pdt := &pc.data[index]

	pdt.Update(pk)

	return pdt
}

func (pc *PanelsMapType) Get() []PanelDataType {
	return pc.data
}

func (pc *PanelsMapType) GetIndex(pk string) int {
	list := pc.data
	for i, pdt := range list {
		if pdt.IsEqualContentString(pk) {
			return i
		}
	}

	return -1
}

func (pc *PanelsMapType) GetDataByIndex(index int) *PanelDataType {
	// If index is out of bounds, return nil
	if index < 0 || index >= len(pc.data) {
		return nil
	}

	return &pc.data[index]
}

func (pc *PanelsMapType) GetSourceCoin(pk string) int64 {
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

func (pc *PanelsMapType) GetTargetCoin(pk string) int64 {
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

func (pc *PanelsMapType) ValidateKey(pk string) bool {
	pko := PanelKeyType{value: pk}
	return pko.Validate()
}

func (pc *PanelsMapType) ValidatePanel(pk string) bool {
	if !pc.ValidateKey(pk) {
		return false
	}

	sid := pc.GetSourceCoin(pk)
	tid := pc.GetTargetCoin(pk)

	if !pc.maps.ValidateId(sid) {
		return false
	}

	if !pc.maps.ValidateId(tid) {
		return false
	}

	return true
}

func (pc *PanelsMapType) ValidateId(id int64) bool {
	return pc.maps.ValidateId(id)
}

func (pc *PanelsMapType) InvalidatePanels() {
	for i := range pc.data {
		p := pc.GetDataByIndex(i)
		p.index = -1
	}
}

func (pc *PanelsMapType) GetOptions() []string {
	return pc.maps.GetOptions()
}

func (pc *PanelsMapType) GetDisplayById(id string) string {
	return pc.maps.GetDisplayById(id)
}

func (pc *PanelsMapType) GetIdByDisplay(id string) string {
	return pc.maps.GetIdByDisplay(id)
}

func (pc *PanelsMapType) GetSymbolById(id string) string {
	return pc.maps.GetSymbolById(id)
}
