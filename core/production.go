//go:build production
// +build production

package core

const MemoryDebug = false

func InitLogger() {
}

func Logln(v ...any) {
}

func Logf(format string, v ...any) {
}
