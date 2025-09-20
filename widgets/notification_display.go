package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
)

type NotificationDisplay struct {
	widget.BaseWidget
	text    *canvas.Text
	padding float32
}

func NewNotificationDisplay() *NotificationDisplay {
	t := canvas.NewText("", JC.TextColor)
	t.Alignment = fyne.TextAlignCenter
	t.TextSize = theme.TextSize()

	w := &NotificationDisplay{
		text:    t,
		padding: 10,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *NotificationDisplay) UpdateText(msg string) {
	maxWidth := w.text.Size().Width
	w.text.Text = JC.TruncateText(msg, maxWidth, w.text.TextSize)
	w.text.Color = JC.TextColor
	w.text.Refresh()
	w.Refresh()
}

func (w *NotificationDisplay) ClearText() {
	JA.StartFadingText(w.text, func() {
		// Clear the text after fade completes
		w.text.Text = ""
		w.text.Color = JC.TextColor
		w.text.Refresh()
		w.Refresh()
	}, nil)
}

func (w *NotificationDisplay) GetText() string {
	return w.text.Text
}

func (w *NotificationDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &notificationDisplayLayout{
		text:      w.text,
		container: w,
	}
}
