package core

import (
	"bytes"
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

func CreateFile(path string, textString string) bool {

	fullPath := strings.TrimLeft(path, "/")
	fileURI, err := storage.ParseURI(fullPath)

	if err != nil {
		Logln("Error getting parsing uri for file:", err)
		return false
	}

	writer, err := storage.Writer(fileURI)
	if err != nil {
		Logf("Error creating writer: %v", err)
		return false
	}

	defer writer.Close()

	_, err = writer.Write([]byte(textString))
	if err != nil {
		Logf("Error writing to file: %v", err)
		return false
	}

	Logf("Successfully created and wrote to file: %s", path)

	return true
}

func FileExists(path string) (bool, error) {

	fileURI, err := storage.ParseURI(path)
	if err != nil {
		Logln("Error getting parsing uri for file:", err)
		return false, err
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		return false, err
	}
	reader.Close()

	return true, nil
}

func BuildPathRelatedToUserDirectory(additionalPath []string) string {

	app := fyne.CurrentApp()

	storageRoot := GetUserDirectory()
	if IsMobile {
		storageRoot = app.Storage().RootURI().Path()
	}

	// Construct the full path
	allPaths := append([]string{storageRoot}, additionalPath...)
	fullpath := filepath.Join(allPaths...)
	uri := storage.NewFileURI(fullpath)

	return uri.String()
}

func GetUserDirectory() string {

	path := []string{}

	// Try using config directory
	configDir, err := os.UserConfigDir()
	if err == nil && len(configDir) > 0 {
		path = append(path, configDir)
	}

	// Try using os home dir
	if len(path) == 0 {
		homeDir, err := os.UserHomeDir()

		if err == nil && len(homeDir) > 0 {
			path = append(path, homeDir)
		}
	}

	// Try using current user home dir
	if len(path) == 0 {
		usr, err := user.Current()
		if err == nil && len(usr.HomeDir) > 0 {
			path = append(path, usr.HomeDir)
		}
	}

	path = append(path, "jxcryptwatcher")

	return filepath.Join(path...)
}

func SaveFile(filename string, data any) bool {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Logln("Error marshaling", filename, err)
		return false
	}
	return CreateFile(BuildPathRelatedToUserDirectory([]string{filename}), string(jsonData))
}

func LoadFile(filename string) (string, bool) {
	fileURI, err := storage.ParseURI(BuildPathRelatedToUserDirectory([]string{filename}))
	if err != nil {
		Logln("Error parsing URI for", filename, err)
		return "", false
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		Logln("Failed to open", filename, err)
		return "", false
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := buffer.ReadFrom(reader); err != nil {
		Logln("Failed to read", filename, err)
		return "", false
	}

	return buffer.String(), true
}
