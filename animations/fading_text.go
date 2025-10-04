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
	JC.UseDispatcher().Submit(func() {
		if fadeAlphas == nil || len(*fadeAlphas) == 0 {
			fadeAlphas = &[]uint8{
				255, 160, 80, 0,
			}
		}

		ticker := time.NewTicker(80 * time.Millisecond)

		go func() {
			defer ticker.Stop()

			for _, alpha := range *fadeAlphas {
				<-ticker.C

				fyne.Do(func() {
					JC.SetTextAlpha(text, alpha)
					text.Refresh()
				})
			}

			if callback != nil {
				fyne.Do(callback)
			}
		}()
	})
}
