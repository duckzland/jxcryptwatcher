package core

import (
	"math/big"
	"strconv"
	"strings"
)

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

func BigFloatNumDecPlaces(f *big.Float) int {
	str := f.Text('f', -1) // full decimal string
	parts := strings.Split(str, ".")
	if len(parts) == 2 {
		return len(strings.TrimRight(parts[1], "0"))
	}
	return 0
}

func ToBigFloat(val float64) *big.Float {
	return new(big.Float).SetPrec(256).SetFloat64(val)
}

func IsNumeric(val string) bool {
	_, err := strconv.Atoi(val)
	return err == nil
}
