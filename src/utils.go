package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

/**
 * Function for extracting number of decimals from a float
 */
func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

/**
 * Helper function for creating file
 */
func createFile(fileName string, textString string) {

	dirPath := filepath.Dir(fileName)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	out, err := os.Create(fileName)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to create file: %w", err)
		log.Fatal(wrappedErr)
	}

	defer out.Close()

	_, err2 := out.WriteString(textString)

	if err2 != nil {
		wrappedErr2 := fmt.Errorf("Failed to write file: %w", err2)
		log.Fatal(wrappedErr2)
	} else {
		log.Printf("Creating new file %s", fileName)
	}

	out.Sync()
	out.Close()

}

/**
 * Helper function for checking if file exists
 */
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func isNumeric(val string) bool {
	_, err := strconv.Atoi(val)
	return err == nil
}

func buildPathRelatedToUserDirectory(additionalPath []string) string {
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

	path = append(path, additionalPath...)
	return filepath.Join(path...)
}
