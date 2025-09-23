package panels

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JM "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type panelContainer struct {
	widget.BaseWidget
	Objects          []fyne.CanvasObject
	layout           *panelGridLayout
	dragPosition     fyne.Position
	dragging         bool
	dragCursorOffset fyne.Position
	fps              time.Duration
	done             chan struct{}
}

func (c *panelContainer) Add(obj fyne.CanvasObject) {
	c.Objects = append(c.Objects, obj)
}

func (c *panelContainer) Remove(obj fyne.CanvasObject) {
	for i, o := range c.Objects {
		if o == obj {
			c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
			break
		}
	}
}

func (c *panelContainer) RemoveByID(uuid string) bool {
	for _, obj := range c.Objects {
		if panel, ok := obj.(*panelDisplay); ok {
			if panel.GetTag() == uuid {
				c.Remove(obj)
				return true
			}
		}
	}
	return false
}

func (c *panelContainer) ForceRefresh() {
	c.layout.Reset()
	c.Refresh()
}

func (c *panelContainer) CreateRenderer() fyne.WidgetRenderer {
	return &panelContainerLayout{
		container: c,
	}
}

func (c *panelContainer) Dragged(ev *fyne.DragEvent) {
	if JM.StatusManager.IsDraggable() {
		return
	}

	// Crash fix nil pointer
	if c == nil || ev == nil {
		return
	}

	c.dragPosition = ev.Position

	if !c.dragging {
		c.dragging = true
		c.dragCursorOffset = ev.Position.Subtract(c.Position())
		c.done = make(chan struct{})

		sourceY := c.Position().Y
		scrollStep := float32(10)
		edgeThreshold := float32(30)

		go func() {
			ticker := time.NewTicker(c.fps)
			defer ticker.Stop()

			for {
				select {
				case <-c.done:
					return
				case <-ticker.C:
					targetY := c.dragPosition.Y - c.dragCursorOffset.Y
					delta := targetY - sourceY
					direction := 0

					if delta < -edgeThreshold {
						direction = -1
					} else if delta > edgeThreshold {
						direction = 1
					} else {
						direction = 0
					}

					if direction != 0 {
						fyne.Do(func() {
							switch direction {
							case -1:
								JM.LayoutManager.ScrollBy(-scrollStep)
							case 1:
								JM.LayoutManager.ScrollBy(scrollStep)
							}
						})
					}
				}
			}
		}()
	}
}

func (c *panelContainer) DragEnd() {
	c.dragging = false
	if c.done != nil {
		close(c.done)
		c.done = nil
	}
}

func (c *panelContainer) UpdatePanelsContent(shouldUpdate func(pdt JT.PanelData) bool) {
	for _, obj := range c.Objects {
		if panel, ok := obj.(*panelDisplay); ok {

			pdt := JT.UsePanelMaps().GetDataByID(panel.GetTag())

			if shouldUpdate != nil && !shouldUpdate(pdt) {
				continue
			}

			panel.updateContent()
		}
	}
}

func NewpanelContainer(
	layout *panelGridLayout,
	Objects []fyne.CanvasObject,
) *panelContainer {
	c := &panelContainer{
		Objects: Objects,
		layout:  layout,
	}

	for _, obj := range c.Objects {
		if panel, ok := obj.(*panelDisplay); ok {
			panel.parent = c
		}
	}

	c.fps = 16 * time.Millisecond
	if JC.IsMobile {
		c.fps = 32 * time.Millisecond
	}

	c.ExtendBaseWidget(c)

	return c
}
