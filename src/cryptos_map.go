package main

import (
	"strconv"
	"strings"
)

type CryptosMap struct {
	data map[string]string
}

func (cm *CryptosMap) Init() {
	cm.data = make(map[string]string)
}

func (cm *CryptosMap) GetOptions() []string {
	m := []string{}

	for _, tk := range cm.data {
		m = append(m, tk)
	}

	return m
}

func (cm *CryptosMap) GetDisplayById(id string) string {
	tid, ok := cm.data[id]
	if !ok {
		return ""
	}

	return tid
}

func (cm *CryptosMap) GetIdByDisplay(tk string) string {
	if isNumeric(tk) {
		return tk
	}

	ntk := strings.Split(tk, "|")
	if len(ntk) > 0 && isNumeric(ntk[0]) {
		_, ok := cm.data[ntk[0]]
		if ok {
			return ntk[0]
		}
	}

	return ""
}

func (cm *CryptosMap) GetSymbolById(id string) string {
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

func (cm *CryptosMap) ValidateId(id int64) bool {
	_, ok := cm.data[strconv.FormatInt(id, 10)]
	return ok
}
