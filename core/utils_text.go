package core

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func DynamicFormatFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', NumDecPlaces(f), 64)
}

func IsAlpha(c color.Color, alpha uint32) bool {
	_, _, _, a := c.RGBA()
	return a == alpha
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
	if str == STRING_EMPTY || sc < 6 {
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
		return STRING_EMPTY
	}

	var b strings.Builder
	b.Grow(maxChars + 3)
	b.WriteString(str[:maxChars])
	b.WriteString(STRING_ELLIPISIS)

	return b.String()
}

func MeasureText(text string, fontSize float32, style fyne.TextStyle) float32 {

	// This is memory hungry!
	// return fyne.MeasureText(text, fontSize, style).Width

	baseFactor := float32(0.58)
	styleMultiplier := float32(1.0)

	if style.Bold {
		styleMultiplier += 0.1
	}
	if style.Italic {
		styleMultiplier += 0.05
	}
	if style.Monospace {
		baseFactor = 0.5
	}

	return fontSize * baseFactor * styleMultiplier * float32(len(text))
}

func FormatShortCurrency(value string) string {
	num, err := strconv.ParseFloat(strings.Replace(value, STRING_DOLLAR, STRING_EMPTY, 1), 64)
	if err != nil {
		return value // fallback if parsing fails
	}

	switch {
	case num >= 1_000_000_000_000:
		return fmt.Sprintf(FMT_SHORT_TRILLION_DOLLAR, num/1_000_000_000_000)
	case num >= 1_000_000_000:
		return fmt.Sprintf(FMT_SHORT_BILLION_DOLLAR, num/1_000_000_000)
	case num >= 1_000_000:
		return fmt.Sprintf(FMT_SHORT_MILLION_DOLLAR, num/1_000_000)
	case num >= 1_000:
		return fmt.Sprintf(FMT_SHORT_THOUSAND_DOLLAR, num/1_000)
	default:
		return fmt.Sprintf(FMT_SHORT_DOLLAR, num)
	}
}

var extractLeadingRegex = regexp.MustCompile(`^\d+`)

func ExtractLeadingNumber(s string) int {
	match := extractLeadingRegex.FindString(s)
	if match == STRING_EMPTY {
		return -1
	}
	num, _ := strconv.Atoi(match)
	return num
}

func SearchableExtractNumber(s string) int {
	parts := strings.SplitN(s, STRING_PIPE, 2)
	if len(parts) < 1 || parts[0] == STRING_EMPTY {
		return -1
	}
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return num
}

func RasterizeText(text string, textStyle fyne.TextStyle, textSize float32, col color.Color, paddingFactor float32, maxPadding float32) (*image.RGBA, fyne.Size) {

	if Window == nil {
		return nil, fyne.Size{}
	}

	scale := Window.Canvas().Scale()
	sampling := SamplingForScale(scale)

	face := UseTheme().GetFontFace(textStyle, textSize, sampling)
	if face == nil {
		return nil, fyne.Size{}
	}

	adv := font.MeasureString(face, text)
	textW := max(adv.Round(), 1)

	padding := float32(math.Ceil(float64(textSize * paddingFactor)))
	if padding > maxPadding {
		padding = maxPadding
	}

	height := float32(math.Ceil(float64(textSize + padding)))

	width := int(math.Ceil(float64(float32(textW) * scale)))

	buf := image.NewRGBA(image.Rect(0, 0, width, int(height)*sampling))

	startX := (width - textW) / 2

	d := &font.Drawer{
		Dst:  buf,
		Src:  image.NewUniform(col),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(startX << 6),
			Y: fixed.Int26_6(int(height-padding) * sampling << 6),
		},
	}
	d.DrawString(text)

	dst := image.NewRGBA(image.Rect(0, 0, width/sampling, int(height)))
	draw.CatmullRom.Scale(dst, dst.Bounds(), buf, buf.Bounds(), draw.Over, nil)

	size := fyne.NewSize(float32(dst.Bounds().Dx()), height)

	return dst, size
}
