package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeInRegistry = JC.NewCancelRegistry()

func StartFadeInBackground(tag string, rect *canvas.Rectangle, duration time.Duration, callback func(), dispatch bool) {

	StopFadeInBackground(tag)

	if !rect.Visible() {
		return
	}

	if dispatch {
		UseAnimationDispatcher().Submit(func() {
			processFadeInBackground(tag, rect, duration, callback)
		})
	} else {
		go processFadeInBackground(tag, rect, duration, callback)
	}
}

func processFadeInBackground(tag string, rect *canvas.Rectangle, duration time.Duration, callback func()) {

	if !rect.Visible() {
		fadeInRegistry.Delete(tag)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeInRegistry.Set(tag, cancel)
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
					rect.Refresh()
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
				rect.Refresh()
			})
		}
	}

	if callback != nil {
		fyne.Do(callback)
	}
}

func StopFadeInBackground(tag string) {
	if cancel, ok := fadeInRegistry.Get(tag); ok {
		cancel()
		fadeInRegistry.Delete(tag)
	}
}
