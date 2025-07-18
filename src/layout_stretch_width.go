package main

import (
	"fyne.io/fyne/v2"
)

type stretchLayout struct {
	Widths []float32
}

func (s *stretchLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	count := len(objects)
	if count == 0 {
		return
	}

	childWidth := size.Width / float32(count)
	curPos := float32(0)

	for i, obj := range objects {
		ww := childWidth
		if len(s.Widths) > i {
			ww = size.Width * s.Widths[i]
		}
		obj.Resize(fyne.NewSize(ww, size.Height))
		obj.Move(fyne.NewPos(curPos, 0))
		curPos = curPos + ww
	}
}

func (s *stretchLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var maxHeight float32
	for _, obj := range objects {
		h := obj.MinSize().Height
		if h > maxHeight {
			maxHeight = h
		}
	}
	return fyne.NewSize(100*float32(len(objects)), maxHeight)
}
