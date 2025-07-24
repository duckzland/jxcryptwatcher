package core

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/canvas"
)

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

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
		log.Printf(
			"%s | Alloc = %v MiB TotalAlloc = %v MiB Sys = %v MiB NumGC = %v",
			title,
			m.Alloc/1024/1024,
			m.TotalAlloc/1024/1024,
			m.Sys/1024/1024,
			m.NumGC,
		)
	}
}

func ExtractLeadingNumber(s string) int {
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(s)
	if match == "" {
		return -1
	}
	num, _ := strconv.Atoi(match)
	return num
}

func ReorderByMatch(arr []string, searchKey string) []string {
	sort.SliceStable(arr, func(i, j int) bool {
		return ExtractLeadingNumber(arr[i]) < ExtractLeadingNumber(arr[j])
	})

	return arr
}

func DynamicFormatFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', NumDecPlaces(f), 64)
}

func SetTextAlpha(text *canvas.Text, alpha uint8) {
	switch c := text.Color.(type) {
	case color.RGBA:
		c.A = alpha
		text.Color = c
	case color.NRGBA:
		c.A = alpha
		text.Color = c
	default:
		// fallback to white with new alpha if type is unknown
		text.Color = color.RGBA{R: 255, G: 255, B: 255, A: alpha}
	}
}
