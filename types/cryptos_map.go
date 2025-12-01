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
	data       map[int64]string
	maps       []string
	searchMaps []string
	mu         sync.RWMutex
}

func parseID(id string) (int64, bool) {
	val, err := strconv.ParseInt(id, 10, 64)
	return val, err == nil
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func (cm *cryptosMapType) Init() {
	cm.mu.Lock()
	cm.data = make(map[int64]string)
	cm.maps = []string{}
	cm.searchMaps = []string{}
	cm.mu.Unlock()
}

func (cm *cryptosMapType) Insert(id string, display string) {
	if intID, ok := parseID(id); ok {
		cm.mu.Lock()
		cm.data[intID] = display
		cm.mu.Unlock()
	}
}

func (cm *cryptosMapType) Hydrate(cache cryptosMapCache) {
	cm.Init()

	cm.mu.Lock()
	if len(cache.Data) != 0 {
		cm.data = make(map[int64]string, len(cache.Data))
		for k, v := range cache.Data {
			if intID, ok := parseID(k); ok {
				cm.data[intID] = v
			}
		}
	}

	if JC.IsMobile {
		if len(cache.Maps) != 0 && len(cache.SearchMaps) != 0 {
			cm.maps = make([]string, len(cache.Maps))
			copy(cm.maps, cache.Maps)

			cm.searchMaps = make([]string, len(cache.SearchMaps))
			copy(cm.searchMaps, cache.SearchMaps)
		}

		if len(cache.Data) == 0 || len(cache.Maps) == 0 || len(cache.SearchMaps) == 0 {
			_ = cm.GetOptions()
		}
	}

	cm.mu.Unlock()
}

func (cm *cryptosMapType) Serialize() cryptosMapCache {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cache := cryptosMapCache{
		Data:       make(map[string]string, len(cm.data)),
		Maps:       make([]string, len(cm.maps)),
		SearchMaps: make([]string, len(cm.searchMaps)),
	}

	for k, v := range cm.data {
		cache.Data[formatID(k)] = v
	}

	if JC.IsMobile {
		copy(cache.Maps, cm.maps)
		copy(cache.SearchMaps, cm.searchMaps)
	}

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

	options := make([]string, 0, len(cm.data))
	lowerSearchMap := make([]string, 0, len(cm.data))

	cm.mu.RLock()
	for _, val := range cm.data {
		options = append(options, val)
		lowerSearchMap = append(lowerSearchMap, strings.ToLower(val))
	}
	cm.mu.RUnlock()

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
	cm.searchMaps = []string{}
	cm.mu.Unlock()
}

func (cm *cryptosMapType) GetDisplayById(id string) string {
	if intID, ok := parseID(id); ok {
		cm.mu.RLock()
		val, ok := cm.data[intID]
		cm.mu.RUnlock()
		if ok {
			return val
		}
	}
	return JC.STRING_EMPTY
}

func (cm *cryptosMapType) GetIdByDisplay(tk string) string {
	if JC.IsNumeric(tk) {
		if intID, ok := parseID(tk); ok {
			cm.mu.RLock()
			_, exists := cm.data[intID]
			cm.mu.RUnlock()
			if exists {
				return tk
			}
		}
	}

	ntk := strings.Split(tk, JC.STRING_PIPE)
	if len(ntk) > 0 && JC.IsNumeric(ntk[0]) {
		if intID, ok := parseID(ntk[0]); ok {
			cm.mu.RLock()
			_, exists := cm.data[intID]
			cm.mu.RUnlock()
			if exists {
				return ntk[0]
			}
		}
	}
	return JC.STRING_EMPTY
}

func (cm *cryptosMapType) GetSymbolById(id string) string {
	if intID, ok := parseID(id); ok {
		cm.mu.RLock()
		val, ok := cm.data[intID]
		cm.mu.RUnlock()
		if ok {
			parts := strings.Split(val, JC.STRING_PIPE)
			if len(parts) == 2 {
				subs := strings.Split(parts[1], " - ")
				if len(subs) >= 2 {
					return subs[0]
				}
			}
		}
	}
	return JC.STRING_EMPTY
}

func (cm *cryptosMapType) GetSymbolByDisplay(tk string) string {
	parts := strings.Split(tk, JC.STRING_PIPE)
	if len(parts) == 2 {
		subs := strings.Split(parts[1], " - ")
		if len(subs) >= 2 {
			return subs[0]
		}
	}
	return JC.STRING_EMPTY
}

func (cm *cryptosMapType) IsEmpty() bool {
	cm.mu.RLock()
	empty := len(cm.data) == 0
	cm.mu.RUnlock()
	return empty
}

func (cm *cryptosMapType) ValidateId(id int64) bool {
	cm.mu.RLock()
	_, ok := cm.data[id]
	cm.mu.RUnlock()
	return ok
}

func NewCryptosMap() *cryptosMapType {
	return &cryptosMapType{}
}

func NewCryptosMapCache() *cryptosMapCache {
	return &cryptosMapCache{}
}
