package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CryptosType struct {
	Values []CryptosValuesType `json:"values"`
}

type CryptosValuesType struct {
	Id       int64
	Name     string
	Symbol   string
	Status   int64
	IsActive int64
}

var CryptosMap map[string]string

func (cp *CryptosValuesType) UnmarshalJSON(data []byte) error {
	var v []interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Fatal(err)
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
func loadCryptos() CryptosType {

	PrintMemUsage("Start loading cryptos.json")

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	io.Copy(b, f)
	f.Close()

	c := CryptosType{}

	err := json.Unmarshal(b.Bytes(), &c)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load cryptos.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Cryptos Loaded")
	}

	PrintMemUsage("End loading cryptos.json")

	return c
}

func checkCryptos() {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	if !exists {
		data := getTickerData()
		createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), data)
	}

	if err != nil {
		log.Fatalln(err)
	}

	populateCryptosMap()
}

func refreshCryptos() {
	data := getTickerData()
	createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), data)
	populateCryptosMap()
}

func populateCryptosMap() {
	PrintMemUsage("Start populating cryptos")
	// Always reset map
	CryptosMap = make(map[string]string)
	c := loadCryptos()

	for _, crypto := range c.Values {

		// Only add crypto that is active at CMC
		if crypto.Status != 0 || crypto.IsActive != 0 {
			tk := createTickerKey(crypto)
			CryptosMap[strconv.FormatInt(crypto.Id, 10)] = tk

			// log.Printf("count %d", len(CryptosOptions))
			// Debug
			// log.Printf(fmt.Sprintf("%d|%s - %s", crypto.Id, crypto.Symbol, crypto.Name))
		}
	}
	PrintMemUsage("End populating cryptos")
}

func createTickerKey(crypto CryptosValuesType) string {
	return fmt.Sprintf("%d|%s - %s", crypto.Id, crypto.Symbol, crypto.Name)
}

func getTickerData() string {

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.DataEndpoint, nil)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetched cryptodata from CMC: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Fetched cryptodata from CMC")
	}

	return string(respBody)
}

func getTickerOptions() []string {
	m := []string{}

	for _, tk := range CryptosMap {
		m = append(m, tk)
	}

	return m
}

func getTickerDisplayById(id string) string {
	tid, ok := CryptosMap[id]
	if !ok {
		return ""
	}

	return tid
}

func getTickerIdByDisplay(tk string) string {
	if isNumeric(tk) {
		return tk
	}

	ntk := strings.Split(tk, "|")
	if len(ntk) > 0 && isNumeric(ntk[0]) {
		_, ok := CryptosMap[ntk[0]]
		if ok {
			return ntk[0]
		}
	}

	return ""
}

func getTickerSymbolById(id string) string {
	tid, ok := CryptosMap[id]
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

func validateCryptoId(id int64) bool {
	_, ok := CryptosMap[strconv.FormatInt(id, 10)]
	return ok
}
