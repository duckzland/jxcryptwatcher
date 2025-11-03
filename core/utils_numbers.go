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

func ToBigString(val string) (*big.Float, bool) {
	return new(big.Float).SetPrec(256).SetString(val)
}

func IsNumeric(val string) bool {
	_, err := strconv.Atoi(val)
	return err == nil
}

func FormatNumberWithCommas(f float64, frac int) string {
	s := strconv.FormatFloat(f, 'f', frac, 64)

	if frac > 0 {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	parts := strings.Split(s, ".")
	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	n := len(intPart)
	var out strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(intPart[i])
	}

	if fracPart != "" {
		out.WriteByte('.')
		out.WriteString(fracPart)
	}
	return out.String()
}
