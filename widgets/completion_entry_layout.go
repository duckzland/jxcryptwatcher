package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	JC "jxwatcher/core"
)

type completionListEntryLayout struct {
	cSize      fyne.Size
	closeSize  fyne.Size
	background *canvas.Rectangle
	listEntry  *completionList
	closeBtn   ActionButton
}

func (l *completionListEntryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width == 0 && size.Height == 0 {
		return
	}

	if size == l.cSize {
		return
	}

	l.cSize = size
	tlPos := fyne.NewPos(0, 0)

	if l.background != nil {
		l.background.CornerRadius = JC.UseTheme().Size(JC.SizePanelBorderRadius)
		if l.background.Size() != size {
			l.background.Resize(size)
		}

		if l.background.Position() != tlPos {
			l.background.Move(tlPos)
		}
	}

	if l.closeBtn != nil {
		if l.closeBtn.Size() != l.closeSize {
			l.closeBtn.Resize(l.closeSize)
		}

		newClosePos := fyne.NewPos(size.Width-l.closeSize.Width, -l.closeSize.Width-3)
		if l.closeBtn.Position() != newClosePos {
			l.closeBtn.Move(newClosePos)
		}
	}

	if l.listEntry != nil {
		if l.listEntry.Size() != size {
			l.listEntry.Resize(size)
		}

		if l.listEntry.Position() != tlPos {
			l.listEntry.Move(tlPos)
		}
	}
}

func (l *completionListEntryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return l.cSize
}

func (l *completionListEntryLayout) Destroy() {

	l.cSize = fyne.Size{}
	l.closeSize = fyne.Size{}

	l.background = nil

	if l.listEntry != nil {
		l.listEntry.Destroy()
		l.listEntry = nil
	}

	l.closeBtn = nil
}
