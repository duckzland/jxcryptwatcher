package animations

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func StartFadeOutBackground(
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
) {
	JC.UseDispatcher().Submit(func() {
		alphaSteps := []uint8{255, 128, 80, 0}
		interval := duration / time.Duration(len(alphaSteps))
		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()

			for _, alpha := range alphaSteps {
				<-ticker.C

				fyne.Do(func() {
					rect.FillColor = JC.SetAlpha(rect.FillColor, float32(alpha))
					rect.Refresh()
				})
			}

			if callback != nil {
				fyne.Do(callback)
			}
		}()
	})
}
