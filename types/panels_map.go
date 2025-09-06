package types

import (
	"sync"

	JC "jxwatcher/core"
)

var BP PanelsMapType

type PanelsMapType struct {
	mu   sync.RWMutex
	Data []*PanelDataType
	Maps *CryptosMapType
}

func (pc *PanelsMapType) Init() {
	pc.Data = []*PanelDataType{}
}

func (pc *PanelsMapType) Set(data []*PanelDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pk := range data {
		if pk.Parent == nil {
			pk.Parent = pc
		}

		pk.Status = JC.STATE_LOADING
	}

	pc.Data = data
}

func (pc *PanelsMapType) SetMaps(maps *CryptosMapType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Maps = maps
}

func (pc *PanelsMapType) Remove(uuid string) bool {

	index := pc.GetIndex(uuid)
	values := pc.Data

	if index < 0 || index >= len(values) {
		return false
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Data = append(values[:index], values[index+1:]...)

	return true

}

func (pc *PanelsMapType) Append(pk string) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.Data == nil {
		pc.Data = []*PanelDataType{}
	}

	pc.Data = append(pc.Data, &PanelDataType{})
	ref := pc.Data[len(pc.Data)-1]

	ref.Init()
	ref.Update(pk)
	ref.Parent = pc
	ref.Status = JC.STATE_FETCHING_NEW

	return ref
}

func (pc *PanelsMapType) Move(uuid string, newIndex int) bool {

	pc.mu.Lock()
	defer pc.mu.Unlock()

	index := -1
	for i, pdt := range pc.Data {
		if pdt.ID == uuid {
			index = i
			break
		}
	}

	if index == -1 || index == newIndex {
		return false
	}

	// Move element by shifting others
	if index < newIndex {
		for i := index; i < newIndex; i++ {
			pc.Data[i], pc.Data[i+1] = pc.Data[i+1], pc.Data[i]
		}
	} else {
		for i := index; i > newIndex; i-- {
			pc.Data[i], pc.Data[i-1] = pc.Data[i-1], pc.Data[i]
		}
	}

	return true
}

func (pc *PanelsMapType) Update(pk string, index int) *PanelDataType {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if index < 0 || index >= len(pc.Data) {
		return nil
	}

	pdt := pc.Data[index]

	pdt.Update(pk)

	return pdt
}

func (pc *PanelsMapType) RefreshData() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for _, pot := range pc.Data {
		pdt := pc.GetData(pot.ID)
		pko := pdt.UsePanelKey()
		mmp := pc.Maps

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

func (pc *PanelsMapType) Get() []*PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]*PanelDataType, len(pc.Data))
	copy(dataCopy, pc.Data)
	return dataCopy
}

func (pc *PanelsMapType) GetIndex(uuid string) int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i, pdt := range pc.Data {
		if pdt.ID == uuid {
			return i
		}
	}

	return -1
}

func (pc *PanelsMapType) GetData(uuid string) *PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i := range pc.Data {
		pdt := pc.GetDataByIndex(i)
		if pdt.ID == uuid {
			return pc.Data[i]
		}
	}

	return nil
}

func (pc *PanelsMapType) GetDataByIndex(index int) *PanelDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if index >= 0 && index < len(pc.Data) {
		return pc.Data[index]
	}

	return nil
}

func (pc *PanelsMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return len(pc.Data) == 0
}

func (pc *PanelsMapType) TotalData() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return len(pc.Data)
}

func (pc *PanelsMapType) ChangeAllStatus(newStatus int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pdt := range pc.Data {
		pdt.Status = newStatus
	}
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
