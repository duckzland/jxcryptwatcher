package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

type PanelsMap struct {
	data binding.UntypedList
	maps CryptosMapType
}

func (pc *PanelsMap) Init() {
	pc.data = binding.NewUntypedList()
}

func (pc *PanelsMap) Set(data binding.UntypedList) {
	pc.data = data
}

func (pc *PanelsMap) SetMaps(maps CryptosMapType) {
	pc.maps = maps
}

func (pc *PanelsMap) Remove(index int) bool {
	values, _ := pc.data.Get()
	if index < 0 || index >= len(values) {
		return false // avoid out-of-bounds
	}

	// Remove item at index
	updated := append(values[:index], values[index+1:]...)
	pc.data.Set(updated)
	return true
}

func (pc *PanelsMap) RemoveByKey(pk string) bool {
	return pc.Remove(pc.GetIndex(pk))
}

func (pc *PanelsMap) Append(pk string) *PanelDataType {

	if !pc.ValidateKey(pk) {
		return nil
	}

	pn := PanelDataType{
		data:   binding.NewString(),
		oldKey: pk,
		parent: pc,
	}

	pn.Set(pk)
	pn.Update(pk)
	pc.data.Append(pn)

	return &pn
}

func (pc *PanelsMap) Insert(pk string, index int) *PanelDataType {

	// Trying to insert to invalid index, exit
	sval, err := pc.data.GetValue(index)
	if err != nil {
		return nil
	}

	pdt, ok := sval.(PanelDataType)
	if !ok {
		return nil
	}

	pdt.Update(pk)

	return &pdt
}

func (pc *PanelsMap) Get() ([]any, error) {
	return pc.data.Get()
}

func (pc *PanelsMap) GenerateKey(source, target, value, decimals string, rate float32) string {
	return fmt.Sprintf("%s-%s-%s-%s|%f", source, target, value, decimals, rate)
}

func (pc *PanelsMap) GenerateKeyFromPanel(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate)
}

func (pc *PanelsMap) GetIndex(pk string) int {
	list, _ := pc.data.Get()
	for i, x := range list {
		pdt, ok := x.(PanelDataType)
		if !ok {
			continue
		}

		if pdt.IsEqualContentString(pk) {
			return i
		}
	}

	return -1
}

func (pc *PanelsMap) GetDataByIndex(index int) *PanelDataType {
	x, err := pc.data.GetValue(index)
	if err != nil {
		return nil
	}

	pdt, ok := x.(PanelDataType)
	if ok {
		return &pdt
	}

	return nil
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

func (pc *PanelsMap) GetIdByDisplay(tk string) string {
	return pc.maps.GetIdByDisplay(tk)
}

func (pc *PanelsMap) GetSymbolById(id string) string {
	return pc.maps.GetSymbolById(id)
}
