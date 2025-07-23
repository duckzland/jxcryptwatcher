package main

import (
	"fmt"
	"strconv"
	"strings"
)

type PanelKeyType struct {
	value string
}

func (p *PanelKeyType) UpdateValue(rate float32) string {
	pkk := strings.Split(p.value, "|")
	p.value = fmt.Sprintf("%s|%f", pkk[0], rate)
	return p.value
}

func (p *PanelKeyType) GenerateKey(source, target, value, decimals string, rate float32) string {
	p.value = fmt.Sprintf("%s-%s-%s-%s|%f", source, target, value, decimals, rate)
	return p.value
}

func (p *PanelKeyType) GenerateKeyFromPanel(panel PanelType, rate float32) string {
	p.value = fmt.Sprintf("%d-%d-%s-%d|%f", panel.Source, panel.Target, dynamicFormatFloatToString(panel.Value), panel.Decimals, rate)
	return p.value
}

func (p *PanelKeyType) Validate() bool {
	pkv := strings.Split(p.value, "|")
	if len(pkv) != 2 {
		return false
	}

	pkt := strings.Split(pkv[0], "-")
	if len(pkt) != 4 {
		return false
	}

	return true
}

func (p *PanelKeyType) GetValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelKeyType) GetValueString() string {

	pkv := strings.Split(p.value, "|")
	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *PanelKeyType) GetSourceCoinInt() int64 {

	source, err := strconv.ParseInt(p.GetSourceCoinString(), 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *PanelKeyType) GetSourceCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 0 {
		return pkv[0]
	}

	return ""
}

func (p *PanelKeyType) GetTargetCoinInt() int64 {

	target, err := strconv.ParseInt(p.GetTargetCoinString(), 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *PanelKeyType) GetTargetCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *PanelKeyType) GetSourceValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetSourceValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelKeyType) GetSourceValueString() string {
	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[2]
	}

	return ""
}

func (p *PanelKeyType) GetDecimalsInt() int64 {
	decimals, err := strconv.ParseInt(p.GetDecimalsString(), 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *PanelKeyType) GetDecimalsString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 2 {
		return pkv[3]
	}

	return ""
}
