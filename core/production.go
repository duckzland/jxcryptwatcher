//go:build production
// +build production

package core

import (
	"io"
	"log"
)

const MemoryDebug = false

func InitLogger() {
	log.SetOutput(io.Discard)
}
