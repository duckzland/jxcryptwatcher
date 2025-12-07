package types

import (
	"fmt"
	"strings"

	JC "jxwatcher/core"

	"github.com/buger/jsonparser"
	"github.com/mozillazg/go-pinyin"
)

type cryptoType struct {
	Id       int64
	Name     string
	Symbol   string
	Status   int64
	IsActive int64
}

// ParseJSON replaces UnmarshalJSON
func (cp *cryptoType) ParseJSON(data []byte) error {
	// Expect array: [id, name, symbol, ..., is_active, status]
	idFloat, err := jsonparser.GetFloat(data, "[0]")
	if err != nil {
		JC.Logln("Invalid 'id' field:", err)
		return err
	}
	nameStr, err := jsonparser.GetString(data, "[1]")
	if err != nil {
		JC.Logln("Invalid 'name' field:", err)
		return err
	}
	symbolStr, err := jsonparser.GetString(data, "[2]")
	if err != nil {
		JC.Logln("Invalid 'symbol' field:", err)
		return err
	}
	isActiveFloat, err := jsonparser.GetFloat(data, "[4]")
	if err != nil {
		JC.Logln("Invalid 'is_active' field:", err)
		return err
	}
	statusFloat, err := jsonparser.GetFloat(data, "[5]")
	if err != nil {
		JC.Logln("Invalid 'status' field:", err)
		return err
	}

	isActive := int64(isActiveFloat)
	status := int64(statusFloat)
	if isActive == 0 || status == 0 {
		return nil
	}

	cp.Id = int64(idFloat)
	cp.Name = cp.sanitizeText(nameStr, true, false, true)
	cp.Symbol = cp.sanitizeText(symbolStr, false, true, false)
	cp.IsActive = isActive
	cp.Status = status

	return nil
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

func (cp *cryptoType) sanitizeText(s string, capitalize bool, allUpper bool, withSpace bool) string {
	if cp.containsCJK(s) {
		var buf strings.Builder
		prevWasCJK := false

		for _, r := range s {
			if r >= 0x4E00 && r <= 0x9FFF {
				a := pinyin.NewArgs()
				result := pinyin.Pinyin(string(r), a)

				if len(result) > 0 {
					syllable := result[0][0]
					if allUpper {
						syllable = strings.ToUpper(syllable)
					} else if capitalize {
						syllable = strings.Title(syllable)
					}

					if withSpace && prevWasCJK {
						buf.WriteRune(' ')
					}
					buf.WriteString(syllable)
					prevWasCJK = true
					continue
				}
			}
			buf.WriteRune(r)
			prevWasCJK = false
		}
		return buf.String()
	}
	return s
}

func (cp *cryptoType) createKey() string {
	symbol := cp.sanitizeText(cp.Symbol, false, true, false)
	name := cp.sanitizeText(cp.Name, true, false, true)
	return fmt.Sprintf("%d|%s - %s", cp.Id, symbol, name)
}
