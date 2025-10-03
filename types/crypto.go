package types

import (
	"encoding/json"
	"fmt"

	JC "jxwatcher/core"
)

type cryptoType struct {
	Id       int64
	Name     string
	Symbol   string
	Status   int64
	IsActive int64
}

func (cp *cryptoType) UnmarshalJSON(data []byte) error {
	var v []interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		JC.Logln(err)
		return err
	}

	if !cp.validate(v) {
		return nil
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

func (cp *cryptoType) validate(v []interface{}) bool {
	if len(v) < 6 {
		JC.Logln("Invalid crypto data length, expected at least 6 fields")
		return false
	}

	// Checking ID
	if _, ok := v[0].(float64); !ok {
		JC.Logln("Invalid 'id' field type in crypto data")
		return false
	}

	// Checking Name
	if _, ok := v[1].(string); !ok {
		JC.Logln("Invalid 'name' field type in crypto data")
		return false
	}

	// Checking Symbol
	if _, ok := v[2].(string); !ok {
		JC.Logln("Invalid 'symbol' field type in crypto data")
		return false
	}

	// Checking Active
	if _, ok := v[4].(float64); !ok {
		JC.Logln("Invalid 'is_active' field type in crypto data")
		return false
	}

	// Checking Name
	if _, ok := v[5].(float64); !ok {
		JC.Logln("Invalid 'status' field type in crypto data")
		return false
	}

	return true
}

func (cp *cryptoType) createKey() string {
	return fmt.Sprintf("%d|%s - %s", cp.Id, cp.Symbol, cp.Name)
}
