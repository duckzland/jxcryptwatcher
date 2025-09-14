package types

import (
	"strconv"
	"strings"
	"sync"

	JC "jxwatcher/core"
)

type CryptosMapCache struct {
	Data map[string]string
	Maps []string
}

type CryptosMapType struct {
	data     sync.Map
	maps     []string
	mapsLock sync.RWMutex
}

func (cm *CryptosMapType) Init() {
	cm.data = sync.Map{}
	cm.mapsLock.Lock()
	cm.maps = []string{}
	cm.mapsLock.Unlock()
}

func (cm *CryptosMapType) Insert(id string, display string) {
	cm.data.Store(id, display)
}

func (cm *CryptosMapType) Hydrate(data map[string]string) {
	cm.Init()
	for id, display := range data {
		cm.Insert(id, display)
	}
	_ = cm.GetOptions()
}

func (cm *CryptosMapType) Serialize() CryptosMapCache {
	cache := CryptosMapCache{
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

func (cm *CryptosMapType) GetOptions() []string {
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
	cm.data.Range(func(_, val any) bool {
		options = append(options, val.(string))
		return true
	})

	cm.mapsLock.Lock()
	cm.maps = options
	cm.mapsLock.Unlock()

	JC.PrintMemUsage("End generating available crypto options")
	return options
}

func (cm *CryptosMapType) SetMaps(m []string) {
	cm.mapsLock.Lock()
	cm.maps = m
	cm.mapsLock.Unlock()
}

func (cm *CryptosMapType) ClearMapCache() {
	cm.mapsLock.Lock()
	cm.maps = []string{}
	cm.mapsLock.Unlock()
}

func (cm *CryptosMapType) GetDisplayById(id string) string {
	if val, ok := cm.data.Load(id); ok {
		return val.(string)
	}
	return ""
}

func (cm *CryptosMapType) GetIdByDisplay(tk string) string {
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

func (cm *CryptosMapType) GetSymbolById(id string) string {
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

func (cm *CryptosMapType) GetSymbolByDisplay(tk string) string {
	parts := strings.Split(tk, "|")
	if len(parts) == 2 {
		subs := strings.Split(parts[1], " - ")
		if len(subs) >= 2 {
			return subs[0]
		}
	}
	return ""
}

func (cm *CryptosMapType) IsEmpty() bool {
	empty := true
	cm.data.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	return empty
}

func (cm *CryptosMapType) ValidateId(id int64) bool {
	_, ok := cm.data.Load(strconv.FormatInt(id, 10))
	return ok
}
