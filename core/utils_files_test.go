package core

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"

	"fyne.io/fyne/v2/app"
)

func init() {
	_ = app.New() // Initializes a dummy Fyne app for storage access
}

type filesNullWriter struct{}

func (n filesNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func filesTurnOffLogs() {
	log.SetOutput(filesNullWriter{})
}

func filesTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestGetUserDirectory(t *testing.T) {
	filesTurnOffLogs()
	defer filesTurnOnLogs()

	dir := GetUserDirectory()
	if dir == "" {
		t.Error("Expected non-empty user directory path")
	}
	if !strings.Contains(dir, "jxcryptwatcher") {
		t.Error("Expected path to include 'jxcryptwatcher'")
	}
}

func TestBuildPathRelatedToUserDirectory(t *testing.T) {
	filesTurnOffLogs()
	defer filesTurnOnLogs()

	path := BuildPathRelatedToUserDirectory([]string{"testfile.json"})
	if !strings.Contains(path, "testfile.json") {
		t.Error("Expected path to include filename")
	}
	if !strings.Contains(path, "jxcryptwatcher") {
		t.Error("Expected path to include user directory")
	}
}

func TestSaveAndLoadFile(t *testing.T) {
	filesTurnOffLogs()
	defer filesTurnOnLogs()

	filename := "test_save_load.json"
	testData := map[string]string{"hello": "world"}

	ok := SaveFileToStorage(filename, testData)
	if !ok {
		t.Fatal("Failed to save file")
	}

	content, ok := LoadFileFromStorage(filename)
	if !ok {
		t.Fatal("Failed to load file")
	}

	var result map[string]string
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal loaded content: %v", err)
	}

	if result["hello"] != "world" {
		t.Errorf("Expected 'hello':'world', got %v", result)
	}
}

func TestCreateFileAndFileExists(t *testing.T) {
	filesTurnOffLogs()
	defer filesTurnOnLogs()

	filename := "test_create_exists.txt"
	fullPath := BuildPathRelatedToUserDirectory([]string{filename})

	ok := CreateFile(fullPath, "sample content")
	if !ok {
		t.Fatal("Failed to create file")
	}

	exists, err := FileExists(fullPath)
	if err != nil {
		t.Fatalf("FileExists returned error: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist")
	}
}

func TestCleanupCreatedFiles(t *testing.T) {
	filesTurnOffLogs()
	defer filesTurnOnLogs()

	files := []string{"test_save_load.json", "test_create_exists.txt"}
	for _, f := range files {
		path := BuildPathRelatedToUserDirectory([]string{f})
		localPath := strings.TrimPrefix(path, "file://")
		if localPath != "" {
			err := os.Remove(localPath)
			if err != nil && !os.IsNotExist(err) {
				t.Errorf("Failed to remove test file %s: %v", f, err)
			}
		}
	}
}
