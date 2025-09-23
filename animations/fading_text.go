package animations

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

func StartFadingText(
	text *canvas.Text,
	callback func(),
	fadeAlphas *[]uint8,
) {
	if fadeAlphas == nil || len(*fadeAlphas) == 0 {
		fadeAlphas = &[]uint8{
			255, 200, 160, 120, 80, 40, 0,
		}
	}

	delay := 80 * time.Millisecond
	if JC.IsMobile {
		delay = 40 * time.Millisecond
	}

	ticker := time.NewTicker(delay)

	go func() {
		defer ticker.Stop()

		for _, alpha := range *fadeAlphas {
			<-ticker.C

			JC.UseDispatcher().Submit(func() {
				fyne.Do(func() {
					JC.SetTextAlpha(text, alpha)
					text.Refresh()
				})
			})
		}

		if callback != nil {
			JC.UseDispatcher().Submit(func() {
				fyne.Do(callback)
			})
		}
	}()
}
