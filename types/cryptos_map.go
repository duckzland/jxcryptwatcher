package types

import (
	"strconv"
	"strings"
	"sync"

	JC "jxwatcher/core"
)

type cryptosMapCache struct {
	Data map[string]string
	Maps []string
}

type cryptosMapType struct {
	data       sync.Map
	maps       []string
	searchMaps []string
	mapsLock   sync.RWMutex
}

func (cm *cryptosMapType) Init() {
	cm.data = sync.Map{}
	cm.mapsLock.Lock()
	cm.maps = []string{}
	cm.searchMaps = []string{}
	cm.mapsLock.Unlock()
}

func (cm *cryptosMapType) Insert(id string, display string) {
	cm.data.Store(id, display)
}

func (cm *cryptosMapType) Hydrate(data map[string]string) {
	cm.Init()
	for id, display := range data {
		cm.Insert(id, display)
	}
	_ = cm.GetOptions()
}

func (cm *cryptosMapType) Serialize() cryptosMapCache {
	cache := cryptosMapCache{
		Data: make(map[string]string),
	}

	cm.mapsLock.RLock()
	cache.Maps = cm.maps
	cm.mapsLock.RUnlock()

	cm.data.Range(func(key, val any) bool {
		k, ok1 := key.(string)
		v, ok2 := val.(string)
		if ok1 && ok2 {
			cache.Data[k] = v
		}
		return true
	})

	return cache
}

func (cm *cryptosMapType) GetOptions() []string {
	JC.PrintMemUsage("Start generating available crypto options")

	cm.mapsLock.RLock()
	if len(cm.maps) != 0 {
		cached := cm.maps
		cm.mapsLock.RUnlock()
		JC.PrintMemUsage("End using cached crypto options")
		return cached
	}
	cm.mapsLock.RUnlock()

	var options []string
	var lowerSearchMap []string

	cm.data.Range(func(_, val any) bool {
		str := val.(string)
		options = append(options, str)
		lowerSearchMap = append(lowerSearchMap, strings.ToLower(str))
		return true
	})

	cm.mapsLock.Lock()
	cm.maps = options
	cm.searchMaps = lowerSearchMap
	cm.mapsLock.Unlock()

	JC.PrintMemUsage("End generating available crypto options")
	return options
}

func (cm *cryptosMapType) GetSearchMap() []string {
	JC.PrintMemUsage("Start retrieving lowercase crypto search map")

	cm.mapsLock.RLock()
	if len(cm.searchMaps) != 0 {
		cached := cm.searchMaps
		cm.mapsLock.RUnlock()
		JC.PrintMemUsage("End using cached crypto search map")
		return cached
	}
	cm.mapsLock.RUnlock()

	// Trigger GetOptions() to generate both maps and searchMap
	_ = cm.GetOptions()

	cm.mapsLock.RLock()
	cached := cm.searchMaps
	cm.mapsLock.RUnlock()

	JC.PrintMemUsage("End retrieving lowercase crypto search map")
	return cached
}

func (cm *cryptosMapType) SetMaps(m []string) {
	cm.mapsLock.Lock()
	cm.maps = m
	cm.mapsLock.Unlock()
}

func (cm *cryptosMapType) ClearMapCache() {
	cm.mapsLock.Lock()
	cm.maps = []string{}
	cm.mapsLock.Unlock()
}

func (cm *cryptosMapType) GetDisplayById(id string) string {
	if val, ok := cm.data.Load(id); ok {
		return val.(string)
	}
	return ""
}

func (cm *cryptosMapType) GetIdByDisplay(tk string) string {
	if JC.IsNumeric(tk) {
		return tk
	}

	ntk := strings.Split(tk, "|")
	if len(ntk) > 0 && JC.IsNumeric(ntk[0]) {
		if _, ok := cm.data.Load(ntk[0]); ok {
			return ntk[0]
		}
	}
	return ""
}

func (cm *cryptosMapType) GetSymbolById(id string) string {
	if val, ok := cm.data.Load(id); ok {
		parts := strings.Split(val.(string), "|")
		if len(parts) == 2 {
			subs := strings.Split(parts[1], " - ")
			if len(subs) >= 2 {
				return subs[0]
			}
		}
	}
	return ""
}

func (cm *cryptosMapType) GetSymbolByDisplay(tk string) string {
	parts := strings.Split(tk, "|")
	if len(parts) == 2 {
		subs := strings.Split(parts[1], " - ")
		if len(subs) >= 2 {
			return subs[0]
		}
	}
	return ""
}

func (cm *cryptosMapType) IsEmpty() bool {
	empty := true
	cm.data.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	return empty
}

func (cm *cryptosMapType) ValidateId(id int64) bool {
	_, ok := cm.data.Load(strconv.FormatInt(id, 10))
	return ok
}

func NewCryptosMap() *cryptosMapType {
	return &cryptosMapType{}
}

func NewCryptosMapCache() *cryptosMapCache {
	return &cryptosMapCache{}
}
