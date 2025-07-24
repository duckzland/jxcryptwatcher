package types

import (
	"strconv"
	"strings"

	JC "jxwatcher/core"
)

type CryptosMapType struct {
	data map[string]string
	maps []string
}

func (cm *CryptosMapType) Init() {
	cm.data = make(map[string]string)
}

func (cm *CryptosMapType) GetOptions() []string {

	JC.PrintMemUsage("Start generating available crypto options")

	if len(cm.maps) != 0 {
		JC.PrintMemUsage("End using cached crypto options")
		return cm.maps
	}

	JC.CryptoOptions = []string{}
	for _, tk := range cm.data {
		cm.maps = append(cm.maps, tk)
	}

	JC.PrintMemUsage("End generating available crypto options")

	return cm.maps
}

func (cm *CryptosMapType) GetDisplayById(id string) string {
	tid, ok := cm.data[id]
	if !ok {
		return ""
	}

	return tid
}

func (cm *CryptosMapType) GetIdByDisplay(tk string) string {
	if JC.IsNumeric(tk) {
		return tk
	}

	ntk := strings.Split(tk, "|")
	if len(ntk) > 0 && JC.IsNumeric(ntk[0]) {
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

func (cm *CryptosMapType) GetSymbolByDisplay(tk string) string {

	ss := strings.Split(tk, "|")
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
