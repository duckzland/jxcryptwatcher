package animations

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeInRegistry = JC.NewCancelRegistry(100)

func StartFadeInBackground(
	tag string,
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
	dispatch bool,
) {
	if cancel, ok := fadeInRegistry.Get(tag); ok {
		cancel()
	}

	if !rect.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeInRegistry.Set(tag, cancel)

	run := func() {
		if !rect.Visible() {
			return
		}

		alphaSteps := []uint8{100, 128, 192, 255}
		interval := duration / time.Duration(len(alphaSteps))
		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()
			defer fadeInRegistry.Delete(tag)
			defer cancel()

			for _, alpha := range alphaSteps {
				select {
				case <-ctx.Done():
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
		}()
	}

	if dispatch {
		UseAnimationDispatcher().Submit(run)
	} else {
		run()
	}
}

func StopFadeInBackground(tag string) {
	if cancel, ok := fadeInRegistry.Get(tag); ok {
		cancel()
		fadeInRegistry.Delete(tag)
	}
}
