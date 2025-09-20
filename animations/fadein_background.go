package animations

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func FadeInBackground(
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
) {
	alphaSteps := []uint8{0, 64, 128, 192, 255}
	if JC.IsMobile {
		alphaSteps = []uint8{0, 128, 255}
	}

	interval := duration / time.Duration(len(alphaSteps))
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for _, alpha := range alphaSteps {
			<-ticker.C
			JC.AnimDispatcher.Submit(func() {
				fyne.Do(func() {
					rect.FillColor = JC.SetAlpha(rect.FillColor, float32(alpha))
					rect.Refresh()
				})
			})
		}

		if callback != nil {
			JC.AnimDispatcher.Submit(callback)
		}
	}()
}
