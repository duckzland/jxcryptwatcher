package animations

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeInRegistry sync.Map

func StartFadeInBackground(
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
) {

	if val, ok := fadeInRegistry.Load(rect); ok {
		if cancel, ok := val.(context.CancelFunc); ok {
			cancel()
		}
	}

	if !rect.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fadeInRegistry.Store(rect, cancel)

	UseAnimationDispatcher().Submit(func() {
		if !rect.Visible() {
			return
		}

		alphaSteps := []uint8{100, 128, 192, 255}
		interval := duration / time.Duration(len(alphaSteps))
		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()
			defer fadeInRegistry.Delete(rect)
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
					ticker.Stop()
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
	})
}
