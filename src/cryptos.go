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

/**
 * Defining struct for Endpoint cryptos.json
 */
type CryptosType struct {
	Values []CryptosValuesType `json:"values"`
}

type CryptosValuesType struct {
	Id     int64
	Name   string
	Symbol string
}

/**
 * Global variables
 */
var Cryptos CryptosType
var CryptosMap map[string]string
var CryptosOptions []string

/**
 * Custom UnmarshalJSON for cryptos.json
 */
func (cp *CryptosValuesType) UnmarshalJSON(data []byte) error {
	var v []interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Fatal(err)
		return err
	}

	cp.Id = int64(v[0].(float64))
	cp.Name = v[1].(string)
	cp.Symbol = v[2].(string)

	return nil
}

/**
 * Loading cryptos.json from CMC
 */
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

/**
 * Load CMC Crypto.json to memory
 */
func loadCryptos() {
	// PrintMemUsage("Start loading cryptos.json")
	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &Cryptos)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load cryptos.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Cryptos Loaded")
	}
	// PrintMemUsage("End loading cryptos.json")
}

/**
 * Helper function to check fo cryptos.json and try to regenerate it when not found
 */
func checkCryptos() {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	if !exists {
		data := getTickerData()
		createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), data)
	}

	if err != nil {
		log.Fatalln(err)
	}

	loadCryptos()
	populateCryptosMap()
}

func refreshCryptos() {
	data := getTickerData()
	createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), data)
	loadCryptos()
	populateCryptosMap()
}

func populateCryptosMap() {
	// Always reset map
	CryptosMap = make(map[string]string)
	CryptosOptions = nil

	for _, crypto := range Cryptos.Values {
		tk := createTickerKey(crypto)
		CryptosMap[strconv.FormatInt(crypto.Id, 10)] = tk
		CryptosOptions = append(CryptosOptions, tk)

		// log.Printf("count %d", len(CryptosOptions))
		// Debug
		// log.Printf(fmt.Sprintf("%d|%s - %s", crypto.Id, crypto.Symbol, crypto.Name))
	}
}

func createTickerKey(crypto CryptosValuesType) string {
	return fmt.Sprintf("%d|%s - %s", crypto.Id, crypto.Symbol, crypto.Name)
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

func validateCryptoId(id int64) bool {
	for _, crypto := range Cryptos.Values {
		if crypto.Id == id {
			return true
		}
	}

	return false
}
