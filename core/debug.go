//go:build !production
// +build !production

package core

import (
	"log"
	"os"
	"runtime"
)

const MemoryDebug = false

func InitLogger() {
	log.SetOutput(os.Stdout) // Disable all logs
}

func Logln(v ...any) {
	log.Println(v...)
}

func Logf(format string, v ...any) {
	log.Printf(format, v...)
}

func PrintMemUsage(title string) {
	if MemoryDebug {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		Logf(
			"%s | Alloc = %v MiB TotalAlloc = %v MiB Sys = %v MiB NumGC = %v",
			title,
			m.Alloc/1024/1024,
			m.TotalAlloc/1024/1024,
			m.Sys/1024/1024,
			m.NumGC,
		)
	}
}
