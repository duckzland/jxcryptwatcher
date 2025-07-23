package core

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const MemoryDebug = false

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
func CreateFile(fileName string, textString string) {

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
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func IsNumeric(val string) bool {
	_, err := strconv.Atoi(val)
	return err == nil
}

func BuildPathRelatedToUserDirectory(additionalPath []string) string {
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

func PrintMemUsage(title string) {
	if MemoryDebug {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Println(title)
		fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
		fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
		fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
		fmt.Printf("\tNumGC = %v\n", m.NumGC)
	}
}

func extractLeadingNumber(s string) int {
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(s)
	if match == "" {
		return -1 // or any fallback value
	}
	num, _ := strconv.Atoi(match)
	return num
}

func ReorderByMatch(arr []string, searchKey string) []string {
	// sort.SliceStable(arr, func(i, j int) bool {
	// 	iMatch := strings.Contains(strings.ToLower(arr[i]), strings.ToLower(searchKey))
	// 	jMatch := strings.Contains(strings.ToLower(arr[j]), strings.ToLower(searchKey))
	// 	return iMatch && !jMatch
	// })

	sort.SliceStable(arr, func(i, j int) bool {
		return extractLeadingNumber(arr[i]) < extractLeadingNumber(arr[j])
	})

	return arr
}

func DynamicFormatFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', NumDecPlaces(f), 64)
}
