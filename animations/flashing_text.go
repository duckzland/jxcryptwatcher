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
	if JC.IsMobile {
		interval = interval / 2
	}

	alphaSequence := make([]uint8, flashes*2)
	for i := range alphaSequence {
		if i%2 == 0 {
			alphaSequence[i] = 200
		} else {
			alphaSequence[i] = 255
		}
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		var lastAlpha uint8 = 0

		for _, alpha := range alphaSequence {
			<-ticker.C

			if alpha != lastAlpha {
				lastAlpha = alpha

				JC.UseDispatcher().Submit(func() {
					fyne.Do(func() {
						JC.SetTextAlpha(text, alpha)
						text.Refresh()
					})
				})
			}
		}
	}()
}
