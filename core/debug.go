//go:build !production
// +build !production

package core

import (
	"log"
	"os"
)

const MemoryDebug = true

func InitLogger() {
	log.SetOutput(os.Stdout) // Disable all logs
}
