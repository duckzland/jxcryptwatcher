package core

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
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

func TruncateText(str string, maxWidth float32, fontSize float32, style fyne.TextStyle) string {
	sc := len(str)
	if str == "" || sc < 6 {
		return str
	}

	styleBits := 0
	if style.Bold {
		styleBits |= 1 << 0
	}

	if style.Italic {
		styleBits |= 1 << 1
	}

	if style.Monospace {
		styleBits |= 1 << 2
	}

	key := int(fontSize)*10 + styleBits
	charWidth, ok := UseCharWidthCache().Get(key)
	if !ok {
		return str
	}

	estimatedWidth := float32(sc) * charWidth
	if estimatedWidth <= maxWidth {
		return str
	}

	ellipsisWidth := 3 * charWidth
	availableWidth := maxWidth - ellipsisWidth
	maxChars := int(availableWidth / charWidth)

	if maxChars <= 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(maxChars + 3)
	b.WriteString(str[:maxChars])
	b.WriteString("...")

	return b.String()
}

func TruncateTextWithEstimation(str string, maxWidth float32, fontSize float32) string {

	const charWidthFactor = 0.6

	ellipsis := "..."
	ellipsisWidth := float32(len([]rune(ellipsis))) * fontSize * charWidthFactor

	runes := []rune(str)
	totalWidth := float32(len(runes)) * fontSize * charWidthFactor

	if totalWidth <= maxWidth {
		return str
	}

	availableWidth := maxWidth - ellipsisWidth
	maxChars := int(availableWidth / (fontSize * charWidthFactor))

	if maxChars <= 0 {
		return ""
	}

	return string(runes[:maxChars]) + ellipsis
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

func SearchableExtractNumber(s string) int {
	parts := strings.SplitN(s, "|", 2)
	if len(parts) < 1 || parts[0] == "" {
		return -1
	}
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return num
}
