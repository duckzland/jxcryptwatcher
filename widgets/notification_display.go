package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
)

var NotificationContainer *notificationDisplay

type notificationDisplay struct {
	widget.BaseWidget
	text     *canvas.Text
	padding  float32
	txtcolor color.Color
}

func NotificationInit() {
	NotificationContainer = NewNotificationDisplay()
}

func NewNotificationDisplay() *notificationDisplay {
	c := JC.MainTheme.Color(theme.ColorNameForeground, theme.VariantDark)

	t := canvas.NewText("", c)
	t.Alignment = fyne.TextAlignCenter
	t.TextSize = JC.NotificationTextSize

	w := &notificationDisplay{
		text:     t,
		padding:  10,
		txtcolor: c,
	}

	w.ExtendBaseWidget(w)

	return w
}

func (w *notificationDisplay) UpdateText(msg string) {
	maxWidth := w.text.Size().Width
	w.text.Text = JC.TruncateText(msg, maxWidth, w.text.TextSize, w.text.TextStyle)
	w.text.Color = w.txtcolor
	w.text.Refresh()
	w.Refresh()
}

func (w *notificationDisplay) ClearText() {
	JA.StartFadingText(w.text, func() {
		w.text.Text = ""
		w.text.Color = w.txtcolor
		w.text.Refresh()
		w.Refresh()
	}, nil)
}

func (w *notificationDisplay) GetText() string {
	return w.text.Text
}

func (w *notificationDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &notificationDisplayLayout{
		text:      w.text,
		container: w,
	}
}
