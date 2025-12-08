package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
)

var fadeTextRegistry = JC.NewCancelRegistry(5)

func StartFadingText(
	tag string,
	text AnimatableText,
	callback func(),
	fadeAlphas *[]uint8,
) {
	// Cancel any existing animation for this tag
	if cancel, ok := fadeTextRegistry.Get(tag); ok {
		cancel()
		fadeTextRegistry.Delete(tag)
	}

	if !text.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeTextRegistry.Set(tag, cancel)

	UseAnimationDispatcher().Submit(func() {
		if !text.Visible() {
			cancel()
			fadeTextRegistry.Delete(tag)
			return
		}

		if fadeAlphas == nil || len(*fadeAlphas) == 0 {
			fadeAlphas = &[]uint8{255, 160, 80, 0}
		}
		interval := 80 * time.Millisecond
		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()
			defer fadeTextRegistry.Delete(tag)

			for _, alpha := range *fadeAlphas {
				select {
				case <-JC.ShutdownCtx.Done():
					cancel()
					return

				case <-ctx.Done():
					// Ensure final state is fully opaque if cancelled
					fyne.Do(func() {
						text.SetAlpha(255)
						text.Refresh()
					})
					return
				case <-ticker.C:
					if !text.Visible() {
						cancel()
						return
					}
					fyne.Do(func() {
						text.SetAlpha(alpha)
						text.Refresh()
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
