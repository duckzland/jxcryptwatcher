package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

// StartFlashingText flashes the given canvas.Text by toggling alpha values
// at a set interval, for the specified number of flashes.
func StartFlashingText(text *canvas.Text, interval time.Duration, visibleColor color.Color, flashes int) {
	go func() {
		for i := 0; i < flashes*2; i++ { // Each flash = 2 toggles
			time.Sleep(interval)
			if i%2 == 0 {
				JC.SetTextAlpha(text, 200)
			} else {
				JC.SetTextAlpha(text, 255)
			}
			fyne.Do(func() {
				text.Refresh()
			})
		}
	}()
}
