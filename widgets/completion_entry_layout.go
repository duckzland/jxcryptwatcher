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

	newCloseSize := fyne.NewSize(closeWidth, closeWidth)
	if closeBtn.Size() != newCloseSize {
		closeBtn.Resize(newCloseSize)
	}

	newClosePos := fyne.NewPos(0, -closeWidth-3)
	if closeBtn.Position() != newClosePos {
		closeBtn.Move(newClosePos)
	}

	newEntrySize := fyne.NewSize(size.Width, height)
	if listEntry.Size() != newEntrySize {
		listEntry.Resize(newEntrySize)
	}

	newEntryPos := fyne.NewPos(0, 0)
	if listEntry.Position() != newEntryPos {
		listEntry.Move(newEntryPos)
	}
}

func (l *completionListEntryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}

	listEntry := objects[0]

	listMin := listEntry.MinSize()

	width := listMin.Width
	height := listMin.Height

	l.cSize = fyne.NewSize(width, height)

	return l.cSize
}
