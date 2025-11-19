package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeTextRegistry = JC.NewCancelRegistry(5)

func StartFadingText(
	tag string,
	text *canvas.Text,
	callback func(),
	fadeAlphas *[]uint8,
) {
	if cancel, ok := fadeTextRegistry.Get(tag); ok {
		cancel()
	}

	if !text.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeTextRegistry.Set(tag, cancel)

	UseAnimationDispatcher().Submit(func() {
		if !text.Visible() {
			return
		}

		if fadeAlphas == nil || len(*fadeAlphas) == 0 {
			fadeAlphas = &[]uint8{255, 160, 80, 0}
		}

		ticker := time.NewTicker(80 * time.Millisecond)

		go func() {
			defer ticker.Stop()
			defer fadeTextRegistry.Delete(tag)
			defer cancel()

			for _, alpha := range *fadeAlphas {
				select {
				case <-ctx.Done():
					if !JC.IsTextAlpha(text, 255) {
						fyne.Do(func() {
							JC.SetTextAlpha(text, 255)
							canvas.Refresh(text)
						})
					}
					return
				case <-ticker.C:
					if !text.Visible() {
						cancel()
						return
					}
					fyne.Do(func() {
						JC.SetTextAlpha(text, alpha)
						canvas.Refresh(text)
					})
				}
			}

			if callback != nil {
				fyne.Do(callback)
			}
		}()
	})
}

func StopFadingText(tag string) {
	if cancel, ok := fadeTextRegistry.Get(tag); ok {
		cancel()
		fadeTextRegistry.Delete(tag)
	}
}
