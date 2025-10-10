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

var notificationContainer *notificationDisplay

type notificationDisplay struct {
	widget.BaseWidget
	text     *canvas.Text
	padding  float32
	txtcolor color.Color
}

func (w *notificationDisplay) UpdateText(msg string) {
	maxWidth := w.text.Size().Width
	txt := JC.TruncateText(msg, maxWidth, w.text.TextSize, w.text.TextStyle)
	if txt == w.text.Text {
		return
	}
	w.text.Text = txt
	w.text.Color = w.txtcolor
	canvas.Refresh(w.text)
	canvas.Refresh(w)
}

func (w *notificationDisplay) ClearText() {
	JA.StartFadingText(w.text, func() {
		w.text.Text = ""
		w.text.Color = w.txtcolor
		canvas.Refresh(w.text)
		canvas.Refresh(w)
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

func UseNotification() *notificationDisplay {
	return notificationContainer
}

func NotificationInit() {
	notificationContainer = NewNotificationDisplay()
}

func NewNotificationDisplay() *notificationDisplay {
	c := JC.UseTheme().GetColor(theme.ColorNameForeground)

	t := canvas.NewText("", c)
	t.Alignment = fyne.TextAlignCenter
	t.TextSize = JC.UseTheme().Size(JC.SizeNotificationText)

	w := &notificationDisplay{
		text:     t,
		padding:  10,
		txtcolor: c,
	}

	w.ExtendBaseWidget(w)

	return w
}
