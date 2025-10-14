package animations

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeRegistry sync.Map // map[*canvas.Rectangle]context.CancelFunc

func StartFadeOutBackground(
	rect *canvas.Rectangle,
	duration time.Duration,
	callback func(),
) {
	// Cancel any existing fade-out for this rectangle
	if val, ok := fadeRegistry.Load(rect); ok {
		if cancel, ok := val.(context.CancelFunc); ok {
			cancel()
		}
	}

	// Create new context for this animation
	ctx, cancel := context.WithCancel(context.Background())
	fadeRegistry.Store(rect, cancel)

	JC.UseDispatcher().Submit(func() {
		alphaSteps := []uint8{255, 128, 80, 0}
		interval := duration / time.Duration(len(alphaSteps))
		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()
			defer fadeRegistry.Delete(rect)

			for _, alpha := range alphaSteps {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
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
