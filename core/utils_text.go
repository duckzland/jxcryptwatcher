package core

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/canvas"
)

func DynamicFormatFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', NumDecPlaces(f), 64)
}

func SetTextAlpha(text *canvas.Text, alpha uint8) {
	switch c := text.Color.(type) {
	case color.RGBA:
		c.A = alpha
		text.Color = c
	case color.NRGBA:
		c.A = alpha
		text.Color = c
	default:
		// fallback to white with new alpha if type is unknown
		text.Color = color.RGBA{R: 255, G: 255, B: 255, A: alpha}
	}
}

func SetAlpha(c color.Color, alpha float32) color.Color {
	r, g, b, _ := c.RGBA()

	return &color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(alpha),
	}
}

func TruncateText(str string, maxWidth float32, fontSize float32) string {
	// Measure full text width with custom font size
	full := canvas.NewText(str, TextColor)
	full.TextSize = fontSize
	size := full.MinSize()

	if size.Width <= maxWidth {
		return str
	}

	// Truncate and add ellipsis
	runes := []rune(str)
	ellipsis := "..."
	for i := len(runes); i > 0; i-- {
		trial := string(runes[:i]) + ellipsis
		tmp := canvas.NewText(trial, TextColor)
		tmp.TextSize = fontSize

		if tmp.MinSize().Width <= maxWidth {
			return trial
		}
	}

	return ""
}

func FormatShortCurrency(value string) string {
	num, err := strconv.ParseFloat(strings.Replace(value, "$", "", 1), 64)
	if err != nil {
		return value // fallback if parsing fails
	}

	switch {
	case num >= 1_000_000_000_000:
		return fmt.Sprintf("$%.2fT", num/1_000_000_000_000)
	case num >= 1_000_000_000:
		return fmt.Sprintf("$%.2fB", num/1_000_000_000)
	case num >= 1_000_000:
		return fmt.Sprintf("$%.2fM", num/1_000_000)
	case num >= 1_000:
		return fmt.Sprintf("$%.2fK", num/1_000)
	default:
		return fmt.Sprintf("$%.2f", num)
	}
}

func ExtractLeadingNumber(s string) int {
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(s)
	if match == "" {
		return -1
	}
	num, _ := strconv.Atoi(match)
	return num
}
