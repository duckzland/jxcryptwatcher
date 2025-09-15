package core

import (
	"runtime"
	"sort"
	"time"

	"github.com/google/uuid"
)

func ReorderByMatch(arr []string, searchKey string) []string {
	sort.SliceStable(arr, func(i, j int) bool {
		return ExtractLeadingNumber(arr[i]) < ExtractLeadingNumber(arr[j])
	})

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
	WorkerManager.PushMessage("notification", msg)
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
