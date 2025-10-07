package apps

import "fyne.io/fyne/v2"

type topBarLayout struct {
	fixedWidth float32
	spacer     float32
	rows       int
	cWidth     float32
	minSize    fyne.Size
	dirty      bool
}

func (s *topBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// Apps is not ready yet!
	if UseLayout() == nil || UseLayout().UseContainer().Size().Width <= 0 || UseLayout().UseContainer().Size().Height <= 0 {
		return
	}

	count := len(objects)
	if count == 0 {
		return
	}

	if s.cWidth == size.Width {
		return
	}

	s.cWidth = size.Width
	s.dirty = true
	s.rows = 1

	// First object fills the rest of the space
	remaining := s.cWidth - (s.fixedWidth+s.spacer)*float32(count-1)

	if remaining < 340 {

		s.rows = 2

		// Layout objects
		curPos := float32(0)
		y := float32(0)
		for i, obj := range objects {
			var w float32
			w = s.fixedWidth
			y = 0

			switch i {
			case 0:
				w = s.cWidth
				y = s.fixedWidth + s.spacer
				curPos = 0

			case 1:
				curPos += remaining/2 + s.spacer

			default:
				curPos += w + s.spacer
			}

			os := fyne.NewSize(w, s.fixedWidth)
			if obj.Size() != os {
				obj.Resize(os)
			}

			op := fyne.NewPos(curPos, y)
			if obj.Position() != op {
				obj.Move(op)
			}
		}

	} else {

		// Layout objects
		curPos := float32(0)
		for i, obj := range objects {
			var w float32
			if i == 0 {
				w = remaining
			} else {
				w = s.fixedWidth
			}

			os := fyne.NewSize(w, s.fixedWidth)
			if obj.Size() != os {
				obj.Resize(os)
			}

			op := fyne.NewPos(curPos, 0)
			if obj.Position() != op {
				obj.Move(op)
			}

			curPos += w + s.spacer
		}
	}
}

func (s *topBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if !s.dirty {
		return s.minSize
	}

	s.dirty = false
	count := len(objects)
	remaining := s.cWidth - (s.fixedWidth+s.spacer)*float32(count-1)
	rows := 1
	maxHeight := s.fixedWidth

	if remaining < 500 {
		rows = 2
	}

	width := s.fixedWidth*float32(len(objects)-1) + 400
	if s.rows > 1 {
		maxHeight = float32(rows)*maxHeight + s.spacer
	}

	s.minSize = fyne.NewSize(width, maxHeight)

	return s.minSize
}
