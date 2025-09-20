package panels

import "fyne.io/fyne/v2"

type panelActionLayout struct {
	margin float32
	height float32
}

func (r *panelActionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := size.Width
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		objSize := obj.MinSize()
		x -= objSize.Width + r.margin
		obj.Move(fyne.NewPos(x, r.margin))
		obj.Resize(objSize)
	}
}

func (r *panelActionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	totalWidth := float32(0)
	maxHeight := float32(0)
	for _, obj := range objects {
		size := obj.MinSize()
		totalWidth += size.Width
		if size.Height > maxHeight {
			maxHeight = size.Height
		}
	}
	return fyne.NewSize(totalWidth, r.height+r.margin)
}
