package core

import (
	"fmt"
	"image"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
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

func SetImageColor(img *image.NRGBA, col color.Color) {
	r, g, b, _ := col.RGBA()
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := img.PixOffset(x, y)
			a := img.Pix[i+3]
			if a > 0 {
				img.Pix[i+0] = uint8(r >> 8)
				img.Pix[i+1] = uint8(g >> 8)
				img.Pix[i+2] = uint8(b >> 8)
			}
		}
	}
}

func SetImageAlpha(img *image.NRGBA, alpha uint8) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := img.PixOffset(x, y)
			orig := img.Pix[i+3]
			img.Pix[i+3] = uint8((uint16(orig) * uint16(alpha)) / 255)
		}
	}
}

func TruncateText(str string, maxWidth float32, fontSize float32, style fyne.TextStyle) string {
	sc := len(str)
	if str == STRING_EMPTY || sc < 6 {
		return str
	}

	estimatedWidth := MeasureText(str, fontSize, style)
	if estimatedWidth <= maxWidth {
		return str
	}

	charWidth := estimatedWidth / float32(sc)
	ellipsisWidth := 3 * charWidth
	availableWidth := maxWidth - ellipsisWidth
	maxChars := int(availableWidth/charWidth) - 3

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

	face := UseTheme().GetFontFace(style, fontSize)
	if face != nil {
		return float32(font.MeasureString(face, text)) / 64.0
	}

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

func RasterizeText(dst *image.NRGBA, text string, textStyle fyne.TextStyle, textSize float32, textAlign fyne.TextAlign, col color.Color) *image.NRGBA {
	face := UseTheme().GetFontFace(textStyle, textSize)
	if face == nil {
		return nil
	}

	metrics := face.Metrics()
	width := max(font.MeasureString(face, text).Round(), 1)
	height := (metrics.Ascent + metrics.Descent).Ceil()

	if dst == nil || width > dst.Bounds().Dx() || height > dst.Bounds().Dy() {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
	} else {
		for i := range dst.Pix {
			dst.Pix[i] = 0
		}
	}

	dx := 0
	switch textAlign {
	case fyne.TextAlignCenter:
		dx = max((dst.Bounds().Dx()-width)/2, 0)
	case fyne.TextAlignTrailing:
		dx = max(dst.Bounds().Dx()-width, 0)
	}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(dx),
			Y: fixed.I(metrics.Ascent.Round()),
		},
	}
	d.DrawString(text)

	return dst
}
