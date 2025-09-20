package widgets

import (
	"math"

	"fyne.io/fyne/v2"
)

type completionListLayout struct {
	itemHeight  float32
	itemVisible int
	lastSize    fyne.Size
	parent      *completionList
}

func (l *completionListLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	if size.Width == 0 && size.Height == 0 {
		return
	}

	l.itemVisible = int(math.Floor(float64(size.Height / l.itemHeight)))

	sWidth := float32(18)

	if len(l.parent.data) <= l.itemVisible {
		sWidth = 0
	}

	cWidth := size.Width - sWidth

	l.parent.contentBox.Resize(fyne.NewSize(cWidth, size.Height))
	l.parent.scrollBox.Resize(fyne.NewSize(sWidth, size.Height))
	l.parent.scrollBox.Move(fyne.NewPos(cWidth, 0))

	if size != l.lastSize {
		if l.itemVisible < 1 {
			l.parent.contentBox.RemoveAll()
			return
		}

		current := len(l.parent.contentBox.Objects)
		l.parent.itemVisible = l.itemVisible

		switch {
		case current < l.itemVisible:
			for i := current; i < l.itemVisible; i++ {
				l.parent.contentBox.Add(NewCompletionText(l.itemHeight, l.parent))
			}

		case current > l.itemVisible:
			for i := current - 1; i >= l.itemVisible; i-- {
				l.parent.contentBox.Remove(l.parent.contentBox.Objects[i])
			}
		}
	}

	l.lastSize = size
	l.parent.prepareForScroll()
	l.parent.refreshContent()
}

func (l *completionListLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return l.lastSize
}
