package animations

import (
	"image/color"
	JC "jxwatcher/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func StartFadingText(
	text *canvas.Text,
	callback func(),
	fadeSteps *[]color.Color,
) {

	go func() {
		if fadeSteps == nil || len(*fadeSteps) == 0 {
			fadeSteps = &[]color.Color{
				color.RGBA{200, 200, 200, 255},
				color.RGBA{150, 150, 150, 255},
				color.RGBA{100, 100, 100, 255},
				color.RGBA{50, 50, 50, 255},
				color.RGBA{0, 0, 0, 255},
			}
		}

		delay := 80 * time.Millisecond
		if JC.IsMobile {
			delay = 40 * time.Millisecond
		}

		for _, c := range *fadeSteps {
			fyne.Do(func() {
				text.Color = c
				text.Refresh()
			})

			time.Sleep(delay)
		}

		if callback != nil {
			callback()
		}
	}()
}
