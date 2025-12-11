package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeOutRegistry = JC.NewCancelRegistry(5)

func StartFadeOutBackground(
	tag string,
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
	dispatch bool,
) {
	if cancel, ok := fadeOutRegistry.Get(tag); ok {
		cancel()
		fadeOutRegistry.Delete(tag)
	}

	if !rect.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeOutRegistry.Set(tag, cancel)

	if dispatch {
		UseAnimationDispatcher().Submit(func() {
			processFadeOutBackground(tag, rect, duration, callback, ctx, cancel)
		})
	} else {
		go processFadeOutBackground(tag, rect, duration, callback, ctx, cancel)
	}

}

func processFadeOutBackground(tag string, rect *canvas.Rectangle, duration time.Duration, callback func(), ctx context.Context, cancel context.CancelFunc) {
	if rect == nil || !rect.Visible() {
		cancel()
		fadeOutRegistry.Delete(tag)
		return
	}

	alphaSteps := []uint8{255, 128, 80, 0}
	interval := duration / time.Duration(len(alphaSteps))
	ticker := time.NewTicker(interval)

	defer ticker.Stop()
	defer cancel()
	defer fadeOutRegistry.Delete(tag)

	for _, alpha := range alphaSteps {
		select {
		case <-JC.ShutdownCtx.Done():
			cancel()
			return

		case <-ctx.Done():
			// Reset to fully opaque if cancelled mid-way
			if !JC.IsAlpha(rect.FillColor, 255) {
				fyne.Do(func() {
					rect.FillColor = JC.SetAlpha(rect.FillColor, 255)
					canvas.Refresh(rect)
				})
			}
			return

		case <-ticker.C:
			if !rect.Visible() {
				cancel()
				return
			}
			fyne.Do(func() {
				rect.FillColor = JC.SetAlpha(rect.FillColor, float32(alpha))
				canvas.Refresh(rect)
			})
		}
	}

	if callback != nil {
		fyne.Do(callback)
	}
}

func StopFadeOutBackground(tag string) {
	if cancel, ok := fadeOutRegistry.Get(tag); ok {
		cancel()
		fadeOutRegistry.Delete(tag)
	}
}
