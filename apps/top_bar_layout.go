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
	if LayoutManager == nil ||
		LayoutManager.ContainerSize().Width <= 0 ||
		LayoutManager.ContainerSize().Height <= 0 ||
		size.Width <= 0 || size.Height <= 0 {
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

	remaining := s.cWidth - (s.fixedWidth+s.spacer)*float32(count-1)

	if remaining < 500 {
		s.rows = 2

		curPos := float32(0)
		y := float32(0)
		for i, obj := range objects {
			w := s.fixedWidth
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

			size := fyne.NewSize(w, s.fixedWidth)
			pos := fyne.NewPos(curPos, y)

			if obj.Size() != size {
				obj.Resize(size)
			}

			if obj.Position() != pos {
				obj.Move(pos)
			}
		}
	} else {
		curPos := float32(0)
		for i, obj := range objects {
			var w float32
			if i == 0 {
				w = remaining
			} else {
				w = s.fixedWidth
			}

			size := fyne.NewSize(w, s.fixedWidth)
			pos := fyne.NewPos(curPos, 0)

			if obj.Size() != size {
				obj.Resize(size)
			}

			if obj.Position() != pos {
				obj.Move(pos)
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
