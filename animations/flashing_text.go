package animations

import (
	"context"
	"image/color"
	"time"

	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
)

var flashRegistry = JC.NewCancelRegistry(50)

func StartFlashingText(
	tag string,
	txt AnimatableText,
	interval time.Duration,
	visibleColor color.Color,
	flashes int,
) {
	if cancel, ok := flashRegistry.Get(tag); ok {
		cancel()
		flashRegistry.Delete(tag)
	}

	if !txt.Visible() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	flashRegistry.Set(tag, cancel)

	r, g, b, _ := visibleColor.RGBA()
	baseR := float64(r >> 8)
	baseG := float64(g >> 8)
	baseB := float64(b >> 8)

	UseAnimationDispatcher().Submit(func() {
		if !txt.Visible() {
			cancel()
			flashRegistry.Delete(tag)
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
			defer func() {
				ticker.Stop()
				flashRegistry.Delete(tag)
			}()

			for _, alpha := range alphaSequence {
				select {
				case <-ctx.Done():
					fyne.Do(func() {
						txt.SetAlpha(255)
						txt.Refresh()
					})
					return
				case <-ticker.C:
					if !txt.Visible() {
						cancel()
						return
					}
					fyne.Do(func() {
						a := float64(alpha) / 255.0
						newCol := color.NRGBA{
							R: uint8(baseR * a),
							G: uint8(baseG * a),
							B: uint8(baseB * a),
							A: 255,
						}
						txt.SetColor(newCol)
						txt.Refresh()
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
