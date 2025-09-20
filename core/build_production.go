//go:build production
// +build production

package core

import "time"

const MemoryDebug = false

func InitLogger() {
}

func Logln(v ...any) {
}

func Logf(format string, v ...any) {
}

func PrintMemUsage(title string) {
}

func PrintExecTime(title string, start time.Time) {
}

func PrintPerfStats(title string, start time.Time) {
}
