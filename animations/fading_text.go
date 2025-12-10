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

	if text == nil || !text.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeTextRegistry.Set(tag, cancel)

	UseAnimationDispatcher().Submit(func() {
		processFadingText(tag, text, callback, fadeAlphas, ctx, cancel)
	})
}

func processFadingText(tag string, text AnimatableText, callback func(), fadeAlphas *[]uint8, ctx context.Context, cancel context.CancelFunc) {
	if text == nil || !text.Visible() {
		cancel()
		fadeTextRegistry.Delete(tag)
		return
	}

	go func(tag string, text AnimatableText, callback func(), fadeAlphas *[]uint8, ctx context.Context, cancel context.CancelFunc) {
		if fadeAlphas == nil || len(*fadeAlphas) == 0 {
			fadeAlphas = &[]uint8{255, 160, 80, 0}
		}

		interval := 80 * time.Millisecond
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer fadeTextRegistry.Delete(tag)

		for _, alpha := range *fadeAlphas {
			select {
			case <-JC.ShutdownCtx.Done():
				cancel()
				return
			case <-ctx.Done():
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
	}(tag, text, callback, fadeAlphas, ctx, cancel)
}

func StopFadingText(tag string) {
	if cancel, ok := fadeTextRegistry.Get(tag); ok {
		cancel()
		fadeTextRegistry.Delete(tag)
	}
}
