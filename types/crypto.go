package types

import (
	"encoding/json"
	"fmt"

	JC "jxwatcher/core"
)

type CryptoType struct {
	Id       int64
	Name     string
	Symbol   string
	Status   int64
	IsActive int64
}

func (cp *CryptoType) UnmarshalJSON(data []byte) error {
	var v []interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		JC.Logln(err)
		return err
	}

	isActive := int64(v[4].(float64))
	status := int64(v[5].(float64))

	if isActive == 0 || status == 0 {
		return nil
	}

	cp.Id = int64(v[0].(float64))
	cp.Name = v[1].(string)
	cp.Symbol = v[2].(string)
	cp.IsActive = int64(v[4].(float64))
	cp.Status = int64(v[5].(float64))

	return nil
}

func (cp *CryptoType) CreateKey() string {
	return fmt.Sprintf("%d|%s - %s", cp.Id, cp.Symbol, cp.Name)
}
