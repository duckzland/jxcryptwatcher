package types

import (
	"sync"

	JC "jxwatcher/core"
)

var panelMapsStorage *panelsMapType = &panelsMapType{}

type panelsMapType struct {
	mu   sync.RWMutex
	data []PanelData
	maps *cryptosMapType
}

func (pc *panelsMapType) Init() {
	pc.SetData([]PanelData{})
}

func (pc *panelsMapType) SetData(data []PanelData) {
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

func (pc *panelsMapType) GetData() []PanelData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]PanelData, len(pc.data))
	copy(dataCopy, pc.data)
	return dataCopy
}

func (pc *panelsMapType) SetMaps(maps *cryptosMapType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.maps = maps
}

func (pc *panelsMapType) GetMaps() *cryptosMapType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.maps
}

func (pc *panelsMapType) HasMaps() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.maps != nil
}

func (pc *panelsMapType) Remove(uuid string) bool {
	index := pc.GetIndex(uuid)
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.data) {
		return false
	}

	pc.data = append(pc.data[:index], pc.data[index+1:]...)
	return true
}

func (pc *panelsMapType) Append(pk string) PanelData {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.data == nil {
		pc.data = []PanelData{}
	}

	ref := &panelDataType{}
	ref.Init()
	ref.Update(pk)
	ref.SetParent(pc)
	ref.SetStatus(JC.STATE_FETCHING_NEW)

	pc.data = append(pc.data, ref)
	return ref
}

func (pc *panelsMapType) Move(uuid string, newIndex int) bool {
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

func (pc *panelsMapType) Update(pk string, index int) PanelData {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.data) {
		return nil
	}

	pdt := pc.data[index]
	pdt.Update(pk)
	return pdt
}

func (pc *panelsMapType) RefreshData() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for _, pot := range pc.data {
		pdt := pc.GetDataByID(pot.GetID())
		if pdt == nil {
			continue
		}

		pko := pdt.UsePanelKey()

		npk := panelType{
			Source:       pko.GetSourceCoinInt(),
			Target:       pko.GetTargetCoinInt(),
			Decimals:     pko.GetDecimalsInt(),
			Value:        pko.GetSourceValueFloat(),
			SourceSymbol: pc.GetSymbolById(pko.GetSourceCoinString()),
			TargetSymbol: pc.GetSymbolById(pko.GetTargetCoinString()),
		}

		pdt.Update(pko.GenerateKeyFromPanel(npk, pko.GetValueFloat()))
		JC.Logln("Panel refreshed: ", pdt.Get())
	}

	return true
}

func (pc *panelsMapType) GetIndex(uuid string) int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i, pdt := range pc.data {
		if pdt.IsID(uuid) {
			return i
		}
	}
	return -1
}

func (pc *panelsMapType) GetDataByID(uuid string) PanelData {
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

func (pc *panelsMapType) GetDataByIndex(index int) PanelData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if index >= 0 && index < len(pc.data) {
		return pc.data[index]
	}
	return nil
}

func (pc *panelsMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data) == 0
}

func (pc *panelsMapType) TotalData() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data)
}

func (pc *panelsMapType) ChangeStatus(newStatus int, shouldChange func(pdt PanelData) bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pdt := range pc.data {
		if shouldChange != nil && !shouldChange(pdt) {
			continue
		}
		pdt.SetStatus(newStatus)
	}
}

func (pc *panelsMapType) Hydrate(data []PanelData) {

	dataLen := pc.TotalData()

	for i := 0; i < dataLen; i++ {
		pdt := pc.GetDataByIndex(i)
		if pdt == nil || i < 0 || i >= len(data) {
			continue
		}

		pkn := data[i]
		pko := pdt.UsePanelKey()

		if !pko.IsConfigMatching(pkn.Get()) {
			continue
		}

		pdt.Set(pkn.Get())
		pdt.SetOldKey(pkn.GetOldKey())
		pdt.SetStatus(pkn.GetStatus())

		if !pdt.HasParent() {
			pdt.SetParent(pc)
		}

		if !pc.ValidatePanel(pdt.Get()) {
			pdt.SetStatus(JC.STATE_BAD_CONFIG)
		}
	}
}

func (pc *panelsMapType) Serialize() []panelDataCache {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var out []panelDataCache
	for _, p := range pc.data {
		if !pc.ValidateKey(p.Get()) {
			continue
		}
		out = append(out, p.Serialize())
	}
	return out
}

func (pc *panelsMapType) UsePanelKey(pk string) *panelKeyType {
	return &panelKeyType{value: pk}
}

func (pc *panelsMapType) ValidateKey(pk string) bool {
	pko := panelKeyType{value: pk}
	return pko.Validate()
}

func (pc *panelsMapType) ValidatePanel(pk string) bool {
	if !pc.ValidateKey(pk) {
		return false
	}

	pko := panelKeyType{value: pk}
	sid := pko.GetSourceCoinInt()
	tid := pko.GetTargetCoinInt()

	return pc.ValidateId(sid) && pc.ValidateId(tid)
}

func (pc *panelsMapType) ValidateId(id int64) bool {
	return pc.maps.ValidateId(id)
}

func (pc *panelsMapType) GetOptions() []string {
	return pc.maps.GetOptions()
}

func (pc *panelsMapType) GetDisplayById(id string) string {
	return pc.maps.GetDisplayById(id)
}

func (pc *panelsMapType) GetIdByDisplay(id string) string {
	return pc.maps.GetIdByDisplay(id)
}

func (pc *panelsMapType) GetSymbolById(id string) string {
	return pc.maps.GetSymbolById(id)
}

func (pc *panelsMapType) GetSymbolByDisplay(id string) string {
	return pc.maps.GetSymbolByDisplay(id)
}

func UsePanelMaps() *panelsMapType {
	return panelMapsStorage
}
