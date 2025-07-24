//go:build production
// +build production

package core

import (
	"io"
	"log"
)

func InitLogger() {
	log.SetOutput(io.Discard) // Disable all logs
}
