package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
)

var fadeTextRegistry = JC.NewCancelRegistry(5)

func StartFadingText(tag string, text AnimatableText, callback func(), fadeAlphas *[]uint8) {

	StopFadingText(tag)

	if text == nil || !text.Visible() {
		return
	}

	UseAnimationDispatcher().Submit(func() {
		processFadingText(tag, text, callback, fadeAlphas)
	})
}

func processFadingText(tag string, text AnimatableText, callback func(), fadeAlphas *[]uint8) {
	if text == nil || !text.Visible() {
		fadeTextRegistry.Delete(tag)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeTextRegistry.Set(tag, cancel)
	defer cancel()
	defer fadeTextRegistry.Delete(tag)

	if fadeAlphas == nil || len(*fadeAlphas) == 0 {
		fadeAlphas = &[]uint8{255, 160, 80, 0}
	}

	interval := 80 * time.Millisecond
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for _, alpha := range *fadeAlphas {
		select {
		case <-JC.ShutdownCtx.Done():
			cancel()
			return
		case <-ctx.Done():
			fyne.Do(func() {
				text.SetAlpha(255)
			})
			return
		case <-ticker.C:
			if !text.Visible() {
				cancel()
				return
			}
			fyne.Do(func() {
				text.SetAlpha(alpha)
			})
		}
	}

	if callback != nil {
		fyne.Do(callback)
	}

}

func StopFadingText(tag string) {
	if cancel, ok := fadeTextRegistry.Get(tag); ok {
		cancel()
		fadeTextRegistry.Delete(tag)
	}
}
