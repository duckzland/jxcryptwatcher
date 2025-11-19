package core

import (
	"bytes"
	"encoding/gob"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	json "github.com/goccy/go-json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

var userConfigDirExists = false

func DeleteFile(path string) bool {
	fullPath := strings.TrimLeft(path, "/")
	fileURI, err := storage.ParseURI(fullPath)

	if err != nil {
		Logln("Error parsing URI for file:", err)
		return false
	}

	if err := storage.Delete(fileURI); err != nil {
		Logf("Error deleting file: %v", err)
		return false
	}

	Logf("Successfully deleted file: %s", path)
	return true
}

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
	fp := filepath.Join(path...)

	return fp
}

func SaveFileToStorage(filename string, data any) bool {
	var content []byte

	switch v := data.(type) {
	case []byte:
		content = v
	default:
		jsonData, err := json.MarshalIndent(v, STRING_EMPTY, "  ")
		if err != nil {
			Logln("Error marshaling", filename, err)
			return false
		}
		content = jsonData
	}

	return CreateFile(BuildPathRelatedToUserDirectory([]string{filename}), string(content))
}

func LoadFileFromStorage(filename string) (string, bool) {
	fileURI, err := storage.ParseURI(BuildPathRelatedToUserDirectory([]string{filename}))
	if err != nil {
		Logln("Error parsing URI for", filename, err)
		return STRING_EMPTY, false
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		Logln("Failed to open", filename, err)
		return STRING_EMPTY, false
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := buffer.ReadFrom(reader); err != nil {
		Logln("Failed to read", filename, err)
		return STRING_EMPTY, false
	}

	return buffer.String(), true
}

func EraseFileFromStorage(filename string) bool {
	path := BuildPathRelatedToUserDirectory([]string{filename})
	return DeleteFile(path)
}

func SaveGobToStorage(filename string, data any) bool {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		Logln("Error encoding GOB", filename, err)
		return false
	}
	return SaveFileToStorage(filename, buf.Bytes())
}

func LoadGobFromStorage(filename string, out interface{}) bool {
	fileURI, err := storage.ParseURI(BuildPathRelatedToUserDirectory([]string{filename}))
	if err != nil {
		Logln("Error parsing URI for", filename, err)
		return false
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		Logln("Failed to open", filename, err)
		return false
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := buffer.ReadFrom(reader); err != nil {
		Logln("Failed to read", filename, err)
		return false
	}

	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(out); err != nil {
		Logln("Failed to decode GOB", filename, err)
		return false
	}

	return true
}
