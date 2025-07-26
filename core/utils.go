package core

import (
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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
)

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

func CreateFile(path string, textString string) bool {

	fullPath := strings.TrimLeft(path, "/")
	fileURI, err := storage.ParseURI(fullPath)

	if err != nil {
		log.Println("Error getting parsing uri for file:", err)
		return false
	}

	writer, err := storage.Writer(fileURI)
	if err != nil {
		log.Printf("Error creating writer: %v", err)
		return false
	}

	defer writer.Close()

	_, err = writer.Write([]byte(textString))
	if err != nil {
		log.Printf("Error writing to file: %v", err)
		return false
	}

	log.Printf("Successfully created and wrote to file: %s", path)

	return true
}

func FileExists(path string) (bool, error) {

	fileURI, err := storage.ParseURI(path)
	if err != nil {
		log.Println("Error getting parsing uri for file:", err)
		return false, err
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		return false, err
	}
	reader.Close()

	return true, nil
}

func IsNumeric(val string) bool {
	_, err := strconv.Atoi(val)
	return err == nil
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

	// log.Println("User Root Directory:", uri.String())
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
