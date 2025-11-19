package animations

import (
	"context"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

var flashRegistry = JC.NewCancelRegistry(100)

func StartFlashingText(
	tag string,
	text *canvas.Text,
	interval time.Duration,
	visibleColor color.Color,
	flashes int,
) {
	if cancel, ok := flashRegistry.Get(tag); ok {
		cancel()
	}

	if !text.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	flashRegistry.Set(tag, cancel)

	r, g, b, _ := visibleColor.RGBA()
	baseR := float64(r >> 8)
	baseG := float64(g >> 8)
	baseB := float64(b >> 8)

	UseAnimationDispatcher().Submit(func() {
		if !text.Visible() {
			return
		}

		alphaSequence := make([]uint8, flashes*2)
		for i := range alphaSequence {
			if i%2 == 0 {
				alphaSequence[i] = 200
			} else {
				alphaSequence[i] = 255
			}
		}

		ticker := time.NewTicker(interval)

		go func() {
			defer ticker.Stop()
			defer flashRegistry.Delete(tag)
			defer cancel()

			for _, alpha := range alphaSequence {
				select {
				case <-ctx.Done():
					if !JC.IsTextAlpha(text, 255) {
						fyne.Do(func() {
							text.Color = visibleColor
							canvas.Refresh(text)
						})
					}
					return
				case <-ticker.C:
					if !text.Visible() {
						cancel()
						return
					}
					fyne.Do(func() {
						a := float64(alpha) / 255.0
						text.Color = color.RGBA{
							R: uint8(baseR * a),
							G: uint8(baseG * a),
							B: uint8(baseB * a),
							A: 255,
						}
						canvas.Refresh(text)
					})
				}
			}
		}()
	})
}

func StopFlashingText(tag string) {
	if cancel, ok := flashRegistry.Get(tag); ok {
		cancel()
		flashRegistry.Delete(tag)
	}
}
