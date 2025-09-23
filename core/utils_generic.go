package core

import (
	"runtime"
	"sort"
	"time"

	"github.com/google/uuid"
)

var hwTotalCPU = runtime.NumCPU()

func TotalCPU() int {
	return hwTotalCPU
}

func ReorderByMatch(arr []string, searchKey string) []string {
	type sortable struct {
		value string
		key   int
	}

	// Precompute keys
	sortables := make([]sortable, len(arr))
	for i, s := range arr {
		sortables[i] = sortable{
			value: s,
			key:   ExtractLeadingNumber(s),
		}
	}

	// Sort using precomputed keys
	sort.SliceStable(sortables, func(i, j int) bool {
		return sortables[i].key < sortables[j].key
	})

	// Rebuild result
	for i, s := range sortables {
		arr[i] = s.value
	}

	return arr
}

func ReorderSearchable(arr []string) []string {
	type sortable struct {
		value string
		key   int
	}

	// Precompute keys
	sortables := make([]sortable, len(arr))
	for i, s := range arr {
		sortables[i] = sortable{
			value: s,
			key:   SearchableExtractNumber(s),
		}
	}

	// Sort using precomputed keys
	sort.SliceStable(sortables, func(i, j int) bool {
		return sortables[i].key < sortables[j].key
	})

	// Rebuild result
	for i, s := range sortables {
		arr[i] = s.value
	}

	return arr
}

func CreateUUID() string {
	id := uuid.New()
	return id.String()
}

func GetMonthBounds(t time.Time) (startUnix, endUnix int64) {
	year, month, _ := t.Date()
	location := t.Location()

	start := time.Date(year, month, 1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 1, 0).Add(-time.Second) // last second of the month

	return start.Unix(), end.Unix()
}

func TraceGoroutines() {
	count := runtime.NumGoroutine()
	Logf("[GOROUTINE TRACE] Active goroutines: %d", count)
}

func Notify(msg string) {
	UseWorker().PushMessage("notification", msg)
}

func EqualStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func EqualIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
