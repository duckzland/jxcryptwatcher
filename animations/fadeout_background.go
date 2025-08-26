package animations

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func FadeOutBackground(rect *canvas.Rectangle, duration time.Duration, callback func()) {
	steps := 30
	interval := duration / time.Duration(steps)

	go func() {
		// Start from full opacity
		origColor := rect.FillColor
		_, _, _, a := origColor.RGBA()
		startAlpha := float32(a) / 257.0 // RGBA returns 0–65535, so divide by 257 to get 0–255

		for i := 0; i <= steps; i++ {
			progress := 1.0 - float32(i)/float32(steps)
			alpha := startAlpha * progress

			time.Sleep(interval)

			fyne.DoAndWait(func() {
				rect.FillColor = JC.SetAlpha(origColor, alpha)
				rect.Refresh()
			})
		}

		if callback != nil {
			callback()
		}
	}()
}
