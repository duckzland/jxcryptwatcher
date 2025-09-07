package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JX "jxwatcher/animations"
	JC "jxwatcher/core"
)

type NotificationDisplayWidget struct {
	text *canvas.Text
}

type NotificationContainer struct {
	*fyne.Container
	Widget *NotificationDisplayWidget
}

func NewNotificationDisplayWidget() *NotificationContainer {
	tw := canvas.NewText("", color.White)
	tw.Alignment = fyne.TextAlignLeading

	widget := &NotificationDisplayWidget{
		text: tw,
	}

	c := container.NewCenter(tw)

	return &NotificationContainer{
		Container: c,
		Widget:    widget,
	}
}

func (nc *NotificationContainer) UpdateText(msg string) {
	// time.Sleep(600 * time.Millisecond) // FYNE layout quirk

	maxWidth := JC.MainLayoutContentWidth - 20

	fyne.Do(func() {
		nc.Widget.text.Text = JC.TruncateText(msg, maxWidth, nc.Widget.text.TextSize)
		nc.Widget.text.Color = color.White
		// nc.Widget.text.Refresh()
		nc.Refresh()
		JC.Logln("Current widget text:", nc.Widget.text.Text, maxWidth, msg)
	})

	// time.Sleep(600 * time.Millisecond)
}

func (nc *NotificationContainer) ClearText() {
	JX.StartFadingText(nc.Widget.text, func() {
		fyne.Do(func() {
			nc.Widget.text.Text = ""
			nc.Widget.text.Color = color.White
			nc.Widget.text.Refresh()
		})
	}, nil)
}

func (nc *NotificationContainer) GetText() string {
	return nc.Widget.text.Text
}
