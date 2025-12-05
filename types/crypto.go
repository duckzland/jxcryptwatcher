package types

import (
	"fmt"
	"strings"

	"github.com/mozillazg/go-pinyin"

	json "github.com/goccy/go-json"

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
	cp.Name = cp.sanitizeText(v[1].(string))
	cp.Symbol = cp.sanitizeText(v[2].(string))
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

func (cp *cryptoType) containsCJK(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x4E00 && r <= 0x9FFF:
			return true
		case r >= 0x3040 && r <= 0x309F:
			return true
		case r >= 0x30A0 && r <= 0x30FF:
			return true
		}
	}
	return false
}

func (cp *cryptoType) sanitizeText(s string) string {
	if cp.containsCJK(s) {
		var out []string
		for _, r := range s {
			if r >= 0x4E00 && r <= 0x9FFF {
				a := pinyin.NewArgs()
				result := pinyin.Pinyin(string(r), a)
				if len(result) > 0 {
					out = append(out, result[0][0])
				}
			} else {
				out = append(out, string(r))
			}
		}

		return strings.Join(out, " ")
	}

	return s
}

func (cp *cryptoType) createKey() string {
	symbol := cp.sanitizeText(cp.Symbol)
	name := cp.sanitizeText(cp.Name)
	return fmt.Sprintf("%d|%s - %s", cp.Id, symbol, name)
}
