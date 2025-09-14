package types

import (
	"sync"

	JC "jxwatcher/core"
)

var BP PanelsMapType = PanelsMapType{}

type PanelsMapType struct {
	mu   sync.RWMutex
	data []*PanelDataType
	maps *CryptosMapType
}

func (pc *PanelsMapType) Init() {
	pc.SetData([]*PanelDataType{})
}

func (pc *PanelsMapType) SetData(data []*PanelDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pk := range data {
		if !pk.HasParent() {
			pk.SetParent(pc)
		}
		pk.SetStatus(JC.STATE_LOADING)
	}
	pc.data = data
}

func (pc *PanelsMapType) GetData() []*PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]*PanelDataType, len(pc.data))
	copy(dataCopy, pc.data)
	return dataCopy
}

func (pc *PanelsMapType) SetMaps(maps *CryptosMapType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.maps = maps
}

func (pc *PanelsMapType) GetMaps() *CryptosMapType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.maps
}

func (pc *PanelsMapType) HasMaps() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.maps != nil
}

func (pc *PanelsMapType) Remove(uuid string) bool {
	index := pc.GetIndex(uuid)
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.data) {
		return false
	}

	pc.data = append(pc.data[:index], pc.data[index+1:]...)
	return true
}

func (pc *PanelsMapType) Append(pk string) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.data == nil {
		pc.data = []*PanelDataType{}
	}

	ref := &PanelDataType{}
	ref.Init()
	ref.Update(pk)
	ref.SetParent(pc)
	ref.SetStatus(JC.STATE_FETCHING_NEW)

	pc.data = append(pc.data, ref)
	return ref
}

func (pc *PanelsMapType) Move(uuid string, newIndex int) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	index := -1
	for i, pdt := range pc.data {
		if pdt.IsID(uuid) {
			index = i
			break
		}
	}

	if index == -1 || index == newIndex {
		return false
	}

	if index < newIndex {
		for i := index; i < newIndex; i++ {
			pc.data[i], pc.data[i+1] = pc.data[i+1], pc.data[i]
		}
	} else {
		for i := index; i > newIndex; i-- {
			pc.data[i], pc.data[i-1] = pc.data[i-1], pc.data[i]
		}
	}

	return true
}

func (pc *PanelsMapType) Update(pk string, index int) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.data) {
		return nil
	}

	pdt := pc.data[index]
	pdt.Update(pk)
	return pdt
}

func (pc *PanelsMapType) RefreshData() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for _, pot := range pc.data {
		pdt := pc.GetDataByID(pot.GetID())
		if pdt == nil {
			continue
		}
		pko := pdt.UsePanelKey()
		mmp := pc.maps

		npk := PanelType{
			Source:       pko.GetSourceCoinInt(),
			Target:       pko.GetTargetCoinInt(),
			Decimals:     pko.GetDecimalsInt(),
			Value:        pko.GetSourceValueFloat(),
			SourceSymbol: mmp.GetSymbolById(pko.GetSourceCoinString()),
			TargetSymbol: mmp.GetSymbolById(pko.GetTargetCoinString()),
		}

		pdt.Update(pko.GenerateKeyFromPanel(npk, float32(pko.GetValueFloat())))
		JC.Logln("Panel refreshed: ", pdt.Get())
	}

	return true
}

func (pc *PanelsMapType) GetIndex(uuid string) int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i, pdt := range pc.data {
		if pdt.IsID(uuid) {
			return i
		}
	}
	return -1
}

func (pc *PanelsMapType) GetDataByID(uuid string) *PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i := range pc.data {
		pdt := pc.GetDataByIndex(i)
		if pdt != nil && pdt.IsID(uuid) {
			return pdt
		}
	}
	return nil
}

func (pc *PanelsMapType) GetDataByIndex(index int) *PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if index >= 0 && index < len(pc.data) {
		return pc.data[index]
	}
	return nil
}

func (pc *PanelsMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data) == 0
}

func (pc *PanelsMapType) TotalData() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data)
}

func (pc *PanelsMapType) ChangeStatus(newStatus int, shouldChange func(pdt *PanelDataType) bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pdt := range pc.data {
		if shouldChange != nil && !shouldChange(pdt) {
			continue
		}
		pdt.SetStatus(newStatus)
	}
}

func (pc *PanelsMapType) Hydrate(data []*PanelDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pk := range data {
		if !pk.HasParent() {
			pk.SetParent(pc)
		}
	}
	pc.data = data
}

func (pc *PanelsMapType) Serialize() []PanelDataCache {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var out []PanelDataCache
	for _, p := range pc.data {
		if !p.IsStatus(JC.STATE_LOADED) {
			continue
		}
		out = append(out, p.Serialize())
	}
	return out
}

func (pc *PanelsMapType) UsePanelKey(pk string) *PanelKeyType {
	return &PanelKeyType{value: pk}
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

	return pc.ValidateId(sid) && pc.ValidateId(tid)
}

func (pc *PanelsMapType) ValidateId(id int64) bool {
	return pc.maps.ValidateId(id)
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

func (pc *PanelsMapType) GetSymbolByDisplay(id string) string {
	return pc.maps.GetSymbolByDisplay(id)
}
