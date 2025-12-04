package core

import (
	"math"
	"runtime"
	"sort"
	"time"

	"github.com/google/uuid"
)

var hwTotalCPU = runtime.NumCPU()

func TotalCPU() int {
	return hwTotalCPU
}

func MaximumThreads(capLimit int) int {
	const throttleFactor = 2

	cpu := TotalCPU()
	max := cpu / throttleFactor

	if max < 1 {
		max = 1
	}
	if max > capLimit {
		max = capLimit
	}
	return max
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
	end := t
	start := t.AddDate(0, 0, -30)

	return start.Unix(), end.Unix()
}

func TraceGoroutines() {
	count := runtime.NumGoroutine()
	Logf("[GOROUTINE TRACE] Active goroutines: %d", count)
}

func Notify(msg string) {
	UseWorker().Push(ACT_NOTIFICATION_PUSH, msg)
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

func SamplingForScale(scale float32) int {
	if scale <= 0 {
		scale = 1.0
	}

	s := min(max(int(math.Ceil(float64(scale))), 2), 4)

	return s
}
