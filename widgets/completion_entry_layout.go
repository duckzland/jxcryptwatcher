package widgets

import "fyne.io/fyne/v2"

type completionListEntryLayout struct {
	cSize fyne.Size
}

func (l *completionListEntryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width == 0 && size.Height == 0 {
		return
	}

	if len(objects) < 2 {
		return
	}

	if size == l.cSize {
		return
	}

	l.cSize = size

	listEntry := objects[0]
	closeBtn := objects[1]

	height := size.Height
	closeWidth := closeBtn.Size().Width

	closeBtn.Resize(fyne.NewSize(closeWidth, closeWidth))
	closeBtn.Move(fyne.NewPos(-closeWidth-2, 2))

	listEntry.Resize(fyne.NewSize(size.Width, height))
	listEntry.Move(fyne.NewPos(0, 0))
}

func (l *completionListEntryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}

	listEntry := objects[0]
	closeBtn := objects[1]

	listMin := listEntry.MinSize()
	closeMin := closeBtn.MinSize()

	width := listMin.Width + closeMin.Width
	height := fyne.Max(listMin.Height, closeMin.Height)

	l.cSize = fyne.NewSize(width, height)

	return l.cSize
}
