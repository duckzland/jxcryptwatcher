//go:build !production
// +build !production

package core

import (
	"log"
	"os"
	"runtime"
	"time"

	"net/http"
	_ "net/http/pprof"
)

const MemoryDebug = true
const PProfDebug = false

func InitLogger() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[JX] ")
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if PProfDebug {
		go func() {
			Logln("Starting pprof server on localhost:6060")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				Logf("pprof server failed: %v", err)
			}
		}()
	}
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

func PrintExecTime(title string, start time.Time) {
	if MemoryDebug {
		elapsed := time.Since(start)
		Logf("%s | ExecTime = %v", title, elapsed)
	}
}

func PrintPerfStats(title string, start time.Time) {
	if MemoryDebug {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		elapsed := time.Since(start)
		Logf(
			"%s | Time=%v | Alloc=%vM | Tot=%vM | GC=%v | Heap=%vM/%vM (%vM idle, %vM rel) | M/F=%v/%v",
			title,
			elapsed,
			m.Alloc/1024/1024,
			m.TotalAlloc/1024/1024,
			m.NumGC,
			m.HeapAlloc/1024/1024,
			m.HeapSys/1024/1024,
			m.HeapIdle/1024/1024,
			m.HeapReleased/1024/1024,
			m.Mallocs,
			m.Frees,
		)
	}
}
