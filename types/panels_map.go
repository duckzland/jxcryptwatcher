package types

import (
	"sync"
)

var BP PanelsMapType

type PanelsMapType struct {
	mu   sync.Mutex
	Data []PanelDataType
	Maps *CryptosMapType
}

func (pc *PanelsMapType) Init() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Data = []PanelDataType{}
}

func (pc *PanelsMapType) Set(data []PanelDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Data = make([]PanelDataType, len(data))

	for i := range data {
		pc.Data[i].Init()
		pc.Data[i].Set(data[i].Get())
		pc.Data[i].Parent = pc
		pc.Data[i].Index = i
	}
}

func (pc *PanelsMapType) SetMaps(maps *CryptosMapType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Maps = maps
}

func (pc *PanelsMapType) Remove(index int) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()

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
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.Data == nil {
		pc.Data = []PanelDataType{}
	}

	pc.Data = append(pc.Data, PanelDataType{})
	ref := &pc.Data[len(pc.Data)-1]

	ref.Init()
	ref.Update(pk)
	ref.Parent = pc
	ref.Index = len(pc.Data) - 1

	return ref
}

func (pc *PanelsMapType) Update(pk string, index int) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.Data) {
		return nil
	}

	pdt := &pc.Data[index]

	pdt.Update(pk)

	return pdt
}

func (pc *PanelsMapType) Get() []PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	return pc.Data
}

func (pc *PanelsMapType) GetIndex(pk string) int {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for i := range pc.Data {
		if pc.Data[i].IsEqualContentString(pk) {
			return i
		}
	}

	return -1
}

func (pc *PanelsMapType) GetDataByIndex(index int) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

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
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for i := range pc.Data {
		pc.Data[i].Index = -1
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
