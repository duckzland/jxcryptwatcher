package animations

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var fadeTextRegistry sync.Map

func StartFadingText(
	text *canvas.Text,
	callback func(),
	fadeAlphas *[]uint8,
) {

	if val, ok := fadeTextRegistry.Load(text); ok {
		if cancel, ok := val.(context.CancelFunc); ok {
			cancel()
		}
	}

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
			defer cancel()

			for _, alpha := range *fadeAlphas {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:

					if !text.Visible() {
						cancel()
						return
					}

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
