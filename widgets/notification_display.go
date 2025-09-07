package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type NotificationDisplayWidget struct {
	widget.BaseWidget
	text    *canvas.Text
	padding float32
}

func NewNotificationDisplayWidget() *NotificationDisplayWidget {
	t := canvas.NewText("", color.White)
	t.Alignment = fyne.TextAlignCenter
	t.TextSize = theme.TextSize()

	w := &NotificationDisplayWidget{
		text:    t,
		padding: 10,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *NotificationDisplayWidget) UpdateText(msg string) {
	maxWidth := w.text.Size().Width
	w.text.Text = JC.TruncateText(msg, maxWidth, w.text.TextSize)
	w.text.Color = color.White
	w.text.Refresh()
	w.Refresh()
}

func (w *NotificationDisplayWidget) ClearText() {
	w.text.Text = ""
	w.text.Color = color.White
	w.text.Refresh()
	w.Refresh()
}

func (w *NotificationDisplayWidget) GetText() string {
	return w.text.Text
}

func (w *NotificationDisplayWidget) CreateRenderer() fyne.WidgetRenderer {
	return &notificationRenderer{
		text:      w.text,
		container: w,
	}
}

type notificationRenderer struct {
	text      *canvas.Text
	container *NotificationDisplayWidget
}

func (r *notificationRenderer) Layout(size fyne.Size) {
	p := r.container.padding
	r.text.Move(fyne.NewPos(p, 0))
	r.text.Resize(fyne.NewSize(size.Width-2*p, size.Height))
}

func (r *notificationRenderer) MinSize() fyne.Size {
	min := r.text.MinSize()
	p := r.container.padding
	return fyne.NewSize(min.Width+2*p, min.Height)
}

func (r *notificationRenderer) Refresh() {
	r.text.Refresh()
}

func (r *notificationRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *notificationRenderer) Destroy() {}
