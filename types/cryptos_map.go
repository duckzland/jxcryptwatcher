package types

import (
	"strconv"
	"strings"
	"sync"

	JC "jxwatcher/core"
)

type CryptosMapType struct {
	data sync.Map
	maps []string
}

func (cm *CryptosMapType) Init() {
	cm.data = sync.Map{}
	cm.maps = []string{}
}

func (cm *CryptosMapType) Insert(id string, display string) {
	cm.data.Store(id, display)
}

func (cm *CryptosMapType) GetOptions() []string {

	JC.PrintMemUsage("Start generating available crypto options")

	if len(cm.maps) != 0 {
		JC.PrintMemUsage("End using cached crypto options")
		return cm.maps
	}

	var options []string

	cm.data.Range(func(_, val any) bool {
		options = append(options, val.(string))
		return true
	})

	cm.maps = options

	JC.PrintMemUsage("End generating available crypto options")

	return cm.maps
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
	cm.data.Range(func(_, _ any) bool {
		return false
	})

	return true
}

func (cm *CryptosMapType) ValidateId(id int64) bool {
	_, ok := cm.data.Load(strconv.FormatInt(id, 10))

	return ok
}

func (cm *CryptosMapType) ClearMapCache() {
	cm.maps = []string{}
}
