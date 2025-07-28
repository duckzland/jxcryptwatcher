package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
)

type NotificationDisplayWidget struct {
	text         *canvas.Text
	msgChan      chan string
	lastActivity time.Time
}

func NewNotificationDisplayWidget(msgChan chan string) *fyne.Container {
	tw := canvas.NewText("", color.White)
	tw.Alignment = fyne.TextAlignLeading

	nd := &NotificationDisplayWidget{
		text:         tw,
		msgChan:      msgChan,
		lastActivity: time.Now(),
	}

	go nd.animateMessages()
	go nd.watchIdleAndClear()

	return container.NewCenter(tw)
}

func (nd *NotificationDisplayWidget) animateMessages() {
	for msg := range nd.msgChan {
		nd.lastActivity = time.Now()

		// Show message instantly
		fyne.Do(func() {
			nd.text.Text = msg
			nd.text.Color = color.White
			nd.text.Refresh()
		})

		time.Sleep(600 * time.Millisecond)
	}
}

func (nd *NotificationDisplayWidget) watchIdleAndClear() {
	for {
		if time.Since(nd.lastActivity) > 6*time.Second && nd.text.Text != "" {

			JC.Logln("Clearing notification display due to inactivity")

			// Clear text
			JA.StartFadingText(nd.text, func() {
				fyne.Do(func() {
					nd.text.Text = ""
					nd.text.Color = color.White // Reset for next message
					nd.text.Refresh()
				})
			}, nil)
		}

		time.Sleep(1000 * time.Millisecond)
	}
}
