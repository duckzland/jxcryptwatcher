package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func DoActionWithNotification(
	showText string,
	completeText string,
	box *widget.Label,
	callback func(),
) {

	go func() {
		callback()

		fyne.Do(func() {
			box.SetText(showText)
		})

		time.Sleep(3 * time.Second)

		fyne.Do(func() {
			box.SetText(completeText)
		})

		time.Sleep(8 * time.Second)

		fyne.Do(func() {
			box.SetText("")
		})
	}()
}
