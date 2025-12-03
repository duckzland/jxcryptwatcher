package widgets

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
)

var notificationContainer *notificationDisplay

type notificationDisplay struct {
	widget.BaseWidget
	text      string
	textSize  float32
	padding   float32
	textStyle fyne.TextStyle
	cSize     fyne.Size
	pSize     fyne.Size
	img       *canvas.Image
	color     color.Color
	txtcolor  color.Color
}

func (w *notificationDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &notificationDisplayLayout{
		padding:   w.padding,
		text:      w.img,
		container: w,
	}
}

func (w *notificationDisplay) MinSize() fyne.Size {
	if w.cSize != fyne.NewSize(0, 0) {
		return w.cSize
	}
	width := JC.MeasureText(w.text, w.textSize, w.textStyle)
	height := w.textSize * 1.35

	w.cSize = fyne.NewSize(width, height)

	return w.cSize
}

func (w *notificationDisplay) Visible() bool {
	return w.BaseWidget.Visible() && w.text != JC.STRING_EMPTY
}

func (w *notificationDisplay) SetText(msg string) {
	maxWidth := w.pSize.Width
	txt := JC.TruncateText(msg, maxWidth, w.textSize, w.textStyle)

	if txt == w.text {
		return
	}

	w.text = txt
	w.cSize = fyne.NewSize(0, 0)

	if w.text == JC.STRING_EMPTY {
		w.Hide()
	} else {
		w.Show()
	}

	w.MinSize()
	w.SetColor(w.txtcolor)
}

func (w *notificationDisplay) GetText() string {
	return w.text
}

func (w *notificationDisplay) ClearText() {
	JA.StartFadingText("n", w, func() {
		w.SetText(JC.STRING_EMPTY)
	}, nil)
}

func (w *notificationDisplay) SetAlpha(a uint8) {
	r, g, b, _ := w.color.RGBA()
	w.color = color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: a,
	}
	w.rasterize()
	w.Refresh()
}

func (w *notificationDisplay) SetColor(col color.Color) {
	w.color = col
	w.rasterize()
	w.Refresh()
}

func (w *notificationDisplay) rasterize() {
	sampling := int(2)
	face := JC.UseTheme().GetFontFace(w.textStyle, w.textSize, sampling)
	if face == nil {
		return
	}

	scale := JC.Window.Canvas().Scale()
	adv := font.MeasureString(face, w.text)
	textW := max(adv.Round(), 0)
	padding := float32(math.Ceil(float64(w.textSize * 0.35)))
	if padding > 4 {
		padding = 4
	}
	height := float32(math.Ceil(float64(w.textSize + padding)))
	width := int(math.Ceil(float64(float32(textW) * scale)))

	buf := image.NewRGBA(image.Rect(0, 0, width, int(height)*sampling))

	startX := (width - textW) / 2

	d := &font.Drawer{
		Dst:  buf,
		Src:  image.NewUniform(w.color),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(startX << 6),
			Y: fixed.Int26_6(int(height-padding) * sampling << 6),
		},
	}
	d.DrawString(w.text)

	dst := image.NewRGBA(image.Rect(0, 0, width/sampling, int(height)))
	draw.CatmullRom.Scale(dst, dst.Bounds(), buf, buf.Bounds(), draw.Over, nil)

	if w.img == nil {
		w.img = canvas.NewImageFromImage(dst)
	} else {
		w.img.Image = dst
	}

	w.img.FillMode = canvas.ImageFillOriginal
	size := fyne.NewSize(float32(dst.Bounds().Dx()), height)

	w.cSize = size
	w.img.SetMinSize(size)
	w.img.Resize(size)
	w.img.Refresh()
}

func UseNotification() *notificationDisplay {
	return notificationContainer
}

func NotificationInit() {
	notificationContainer = NewNotificationDisplay()
}

func NewNotificationDisplay() *notificationDisplay {
	c := JC.UseTheme().GetColor(theme.ColorNameForeground)
	w := &notificationDisplay{
		text:      JC.STRING_EMPTY,
		color:     c,
		textSize:  JC.UseTheme().Size(JC.SizeNotificationText),
		textStyle: fyne.TextStyle{Bold: false},
		padding:   10,
		txtcolor:  c,
		img:       canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 0, 0))),
	}

	w.ExtendBaseWidget(w)

	return w
}
