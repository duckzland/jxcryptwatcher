package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeInRegistry = JC.NewCancelRegistry(50)

func StartFadeInBackground(
	tag string,
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
	dispatch bool,
) {
	if cancel, ok := fadeInRegistry.Get(tag); ok {
		cancel()
		fadeInRegistry.Delete(tag)
	}

	if !rect.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeInRegistry.Set(tag, cancel)

	if dispatch {
		UseAnimationDispatcher().Submit(func() {
			processFadeInBackground(tag, rect, duration, callback, ctx, cancel)
		})
	} else {
		processFadeInBackground(tag, rect, duration, callback, ctx, cancel)
	}
}

func processFadeInBackground(tag string, rect *canvas.Rectangle, duration time.Duration, callback func(), ctx context.Context, cancel context.CancelFunc) {

	if !rect.Visible() {
		cancel()
		fadeInRegistry.Delete(tag)
		return
	}

	go func(tag string, rect *canvas.Rectangle, duration time.Duration, callback func(), ctx context.Context, cancel context.CancelFunc) {
		defer cancel()
		defer fadeInRegistry.Delete(tag)

		alphaSteps := []uint8{100, 128, 192, 255}
		interval := duration / time.Duration(len(alphaSteps))

		for _, alpha := range alphaSteps {
			select {
			case <-JC.ShutdownCtx.Done():
				cancel()
				return

			case <-ctx.Done():
				if !JC.IsAlpha(rect.FillColor, 255) {
					fyne.Do(func() {
						rect.FillColor = JC.SetAlpha(rect.FillColor, 255)
						canvas.Refresh(rect)
					})
				}
				return
			default:
				if !rect.Visible() {
					cancel()
					return
				}
				time.Sleep(interval)
				fyne.Do(func() {
					rect.FillColor = JC.SetAlpha(rect.FillColor, float32(alpha))
					canvas.Refresh(rect)
				})
			}
		}

		if callback != nil {
			fyne.Do(callback)
		}
	}(tag, rect, duration, callback, ctx, cancel)
}

func StopFadeInBackground(tag string) {
	if cancel, ok := fadeInRegistry.Get(tag); ok {
		cancel()
		fadeInRegistry.Delete(tag)
	}
}
