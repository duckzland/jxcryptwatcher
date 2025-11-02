package types

import (
	"strconv"
	"strings"
	"sync"
	"time"

	JC "jxwatcher/core"
)

type cryptosMapCache struct {
	Data       map[string]string
	Maps       []string
	SearchMaps []string
}

type cryptosMapType struct {
	data       sync.Map
	maps       []string
	searchMaps []string
	mu         sync.RWMutex
}

func (cm *cryptosMapType) Init() {
	cm.data = sync.Map{}
	cm.mu.Lock()
	cm.maps = []string{}
	cm.searchMaps = []string{}
	cm.mu.Unlock()
}

func (cm *cryptosMapType) Insert(id string, display string) {
	cm.data.Store(id, display)
}

func (cm *cryptosMapType) Hydrate(cache cryptosMapCache) {
	cm.Init()

	for id, display := range cache.Data {
		cm.Insert(id, display)
	}

	if len(cache.Maps) != 0 && len(cache.SearchMaps) != 0 {
		cm.mu.Lock()
		cm.maps = cache.Maps
		cm.searchMaps = cache.SearchMaps
		cm.mu.Unlock()
	} else {
		_ = cm.GetOptions()
	}
}

func (cm *cryptosMapType) Serialize() cryptosMapCache {
	cache := cryptosMapCache{
		Data: make(map[string]string),
	}

	cm.mu.RLock()
	cache.Maps = cm.maps
	cache.SearchMaps = cm.searchMaps
	cm.mu.RUnlock()

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
	cm.mu.RLock()
	if len(cm.maps) != 0 {
		cached := cm.maps
		cm.mu.RUnlock()
		return cached
	}
	cm.mu.RUnlock()

	JC.PrintPerfStats("Generating crypto options", time.Now())

	var options []string
	var lowerSearchMap []string

	cm.data.Range(func(_, val any) bool {
		str := val.(string)
		options = append(options, str)
		lowerSearchMap = append(lowerSearchMap, strings.ToLower(str))
		return true
	})

	cm.mu.Lock()
	cm.maps = options
	cm.searchMaps = lowerSearchMap
	cm.mu.Unlock()

	return options
}

func (cm *cryptosMapType) GetSearchMap() []string {

	cm.mu.RLock()
	if len(cm.searchMaps) != 0 {
		cached := cm.searchMaps
		cm.mu.RUnlock()
		return cached
	}
	cm.mu.RUnlock()

	// Trigger GetOptions() to generate both maps and searchMap
	_ = cm.GetOptions()

	cm.mu.RLock()
	cached := cm.searchMaps
	cm.mu.RUnlock()

	return cached
}

func (cm *cryptosMapType) SetMaps(m []string) {
	cm.mu.Lock()
	cm.maps = m
	cm.mu.Unlock()
}

func (cm *cryptosMapType) ClearMapCache() {
	cm.mu.Lock()
	cm.maps = []string{}
	cm.mu.Unlock()
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
