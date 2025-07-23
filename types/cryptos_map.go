package types

import (
	"jxwatcher/core"
	"strconv"
	"strings"
)

type CryptosMapType struct {
	data map[string]string
}

func (cm *CryptosMapType) Init() {
	cm.data = make(map[string]string)
}

func (cm *CryptosMapType) GetOptions() []string {
	m := []string{}

	for _, tk := range cm.data {
		m = append(m, tk)
	}

	return m
}

func (cm *CryptosMapType) GetDisplayById(id string) string {
	tid, ok := cm.data[id]
	if !ok {
		return ""
	}

	return tid
}

func (cm *CryptosMapType) GetIdByDisplay(tk string) string {
	if core.IsNumeric(tk) {
		return tk
	}

	ntk := strings.Split(tk, "|")
	if len(ntk) > 0 && core.IsNumeric(ntk[0]) {
		_, ok := cm.data[ntk[0]]
		if ok {
			return ntk[0]
		}
	}

	return ""
}

func (cm *CryptosMapType) GetSymbolById(id string) string {
	tid, ok := cm.data[id]
	if !ok {
		return ""
	}

	ss := strings.Split(tid, "|")
	if len(ss) != 2 {
		return ""
	}

	sss := strings.Split(ss[1], " - ")
	if len(sss) < 2 {
		return ""
	}

	return sss[0]
}

func (cm *CryptosMapType) ValidateId(id int64) bool {
	_, ok := cm.data[strconv.FormatInt(id, 10)]
	return ok
}
