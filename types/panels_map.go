package types

import (
	"fyne.io/fyne/v2/data/binding"
)

var BP PanelsMapType

type PanelsMapType struct {
	Data []PanelDataType
	Maps CryptosMapType
}

func (pc *PanelsMapType) Init() {
	pc.Data = []PanelDataType{}
}

func (pc *PanelsMapType) Set(data []PanelDataType) {
	pc.Data = data
}

func (pc *PanelsMapType) SetMaps(maps CryptosMapType) {
	pc.Maps = maps
}

func (pc *PanelsMapType) Remove(index int) bool {
	values := pc.Data
	if index < 0 || index >= len(values) {
		return false
	}

	pc.Data = append(values[:index], values[index+1:]...)

	return true
}

func (pc *PanelsMapType) RemoveByKey(pk string) bool {
	return pc.Remove(pc.GetIndex(pk))
}

func (pc *PanelsMapType) Append(pk string) *PanelDataType {

	if pc.Data == nil {
		pc.Data = []PanelDataType{}
	}

	pn := PanelDataType{
		Data: binding.NewString(),
		// OldKey: pk,
		Parent: pc,
		Index:  -1,
	}

	pn.Update(pk)
	pc.Data = append(pc.Data, pn)

	return &pn
}

func (pc *PanelsMapType) Update(pk string, index int) *PanelDataType {

	if index < 0 || index >= len(pc.Data) {
		return nil
	}

	pdt := &pc.Data[index]

	pdt.Update(pk)

	return pdt
}

func (pc *PanelsMapType) Get() []PanelDataType {
	return pc.Data
}

func (pc *PanelsMapType) GetIndex(pk string) int {
	list := pc.Data
	for i, pdt := range list {
		if pdt.IsEqualContentString(pk) {
			return i
		}
	}

	return -1
}

func (pc *PanelsMapType) GetDataByIndex(index int) *PanelDataType {
	// If index is out of bounds, return nil
	if index < 0 || index >= len(pc.Data) {
		return nil
	}

	return &pc.Data[index]
}

func (pc *PanelsMapType) UsePanelKey(pk string) *PanelKeyType {
	pko := PanelKeyType{value: pk}
	return &pko
}

func (pc *PanelsMapType) ValidateKey(pk string) bool {
	pko := PanelKeyType{value: pk}
	return pko.Validate()
}

func (pc *PanelsMapType) ValidatePanel(pk string) bool {
	if !pc.ValidateKey(pk) {
		return false
	}

	pko := PanelKeyType{value: pk}
	sid := pko.GetSourceCoinInt()
	tid := pko.GetTargetCoinInt()

	if !pc.ValidateId(sid) {
		return false
	}

	if !pc.ValidateId(tid) {
		return false
	}

	return true
}

func (pc *PanelsMapType) ValidateId(id int64) bool {
	return pc.Maps.ValidateId(id)
}

func (pc *PanelsMapType) InvalidatePanels() {
	for i := range pc.Data {
		p := pc.GetDataByIndex(i)
		p.Index = -1
	}
}

func (pc *PanelsMapType) GetOptions() []string {
	return pc.Maps.GetOptions()
}

func (pc *PanelsMapType) GetDisplayById(id string) string {
	return pc.Maps.GetDisplayById(id)
}

func (pc *PanelsMapType) GetIdByDisplay(id string) string {
	return pc.Maps.GetIdByDisplay(id)
}

func (pc *PanelsMapType) GetSymbolById(id string) string {
	return pc.Maps.GetSymbolById(id)
}

func (pc *PanelsMapType) GetSymbolByDisplay(id string) string {
	return pc.Maps.GetSymbolByDisplay(id)
}
