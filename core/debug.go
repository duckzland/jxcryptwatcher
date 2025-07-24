//go:build !production
// +build !production

package core

import (
	"log"
	"os"
)

func InitLogger() {
	log.SetOutput(os.Stdout) // Disable all logs
}
