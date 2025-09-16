package animations

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func StartFlashingText(
	text *canvas.Text,
	interval time.Duration,
	visibleColor color.Color,
	flashes int,
) {
	go func() {

		if JC.IsMobile {
			interval = interval / 2
		}

		for i := 0; i < flashes*2; i++ {
			time.Sleep(interval)
			fyne.Do(func() {
				if i%2 == 0 {
					JC.SetTextAlpha(text, 200)
				} else {
					JC.SetTextAlpha(text, 255)
				}
				text.Refresh()
			})
		}
	}()
}
