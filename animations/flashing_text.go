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

	r, g, b, _ := visibleColor.RGBA()
	baseR := float64(r >> 8)
	baseG := float64(g >> 8)
	baseB := float64(b >> 8)

	JC.UseDispatcher().Submit(func() {
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

			for _, alpha := range alphaSequence {
				<-ticker.C

				fyne.Do(func() {
					a := float64(alpha) / 255.0
					text.Color = color.RGBA{
						R: uint8(baseR * a),
						G: uint8(baseG * a),
						B: uint8(baseB * a),
						A: 255,
					}
					canvas.Refresh(text)
				})
			}
		}()
	})
}
