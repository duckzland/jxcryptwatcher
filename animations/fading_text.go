package animations

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeTextRegistry sync.Map // map[*canvas.Text]context.CancelFunc

func StartFadingText(
	text *canvas.Text,
	callback func(),
	fadeAlphas *[]uint8,
) {
	// Cancel any existing fade for this text
	if val, ok := fadeTextRegistry.Load(text); ok {
		if cancel, ok := val.(context.CancelFunc); ok {
			cancel()
		}
	}

	// Create new context for this animation
	ctx, cancel := context.WithCancel(context.Background())
	fadeTextRegistry.Store(text, cancel)

	JC.UseDispatcher().Submit(func() {
		if fadeAlphas == nil || len(*fadeAlphas) == 0 {
			fadeAlphas = &[]uint8{255, 160, 80, 0}
		}

		ticker := time.NewTicker(80 * time.Millisecond)

		go func() {
			defer ticker.Stop()
			defer fadeTextRegistry.Delete(text)

			for _, alpha := range *fadeAlphas {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
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
