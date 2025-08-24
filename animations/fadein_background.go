package animations

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func FadeInBackground(rect *canvas.Rectangle, duration time.Duration, callback func()) {
	steps := 30
	interval := duration / time.Duration(steps)

	go func() {
		_, _, _, a := rect.FillColor.RGBA()
		origAlpha := float32(a) / 257.0

		for i := 0; i <= steps; i++ {
			progress := float32(i) / float32(steps)

			time.Sleep(interval)

			fyne.Do(func() {
				rect.FillColor = JC.SetAlpha(rect.FillColor, origAlpha*progress)
				rect.Refresh()
			})
		}

		if callback != nil {
			callback()
		}
	}()
}
