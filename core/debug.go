//go:build !production
// +build !production

package core

import (
	"log"
	"os"
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
