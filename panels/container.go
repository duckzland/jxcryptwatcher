package panels

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JM "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type panelGridRenderer struct {
	container *PanelGridContainer
}

func (r *panelGridRenderer) Layout(size fyne.Size) {
	r.container.layout.Layout(r.container.Objects, size)
}

func (r *panelGridRenderer) MinSize() fyne.Size {
	return r.container.layout.MinSize(r.container.Objects)
}

func (r *panelGridRenderer) Refresh() {
	JC.MainDebouncer.Call("panel_container_refresh", 10*time.Millisecond, func() {
		fyne.Do(func() {
			r.Layout(r.container.Size())
		})
	})
}

func (r *panelGridRenderer) Objects() []fyne.CanvasObject {
	return r.container.Objects
}

func (r *panelGridRenderer) Destroy() {
	r.container.dragging = false
}

type PanelGridContainer struct {
	widget.BaseWidget
	Objects          []fyne.CanvasObject
	layout           *PanelGridLayout
	dragPosition     fyne.Position
	dragging         bool
	dragCursorOffset fyne.Position
	fps              time.Duration
}

func (c *PanelGridContainer) Add(obj fyne.CanvasObject) {
	c.Objects = append(c.Objects, obj)
}

func (c *PanelGridContainer) Remove(obj fyne.CanvasObject) {
	for i, o := range c.Objects {
		if o == obj {
			c.Objects = append(c.Objects[:i], c.Objects[i+1:]...)
			break
		}
	}
}

func (c *PanelGridContainer) ForceRefresh() {
	c.layout.Reset()
	c.Refresh()
}

func (c *PanelGridContainer) CreateRenderer() fyne.WidgetRenderer {
	return &panelGridRenderer{
		container: c,
	}
}

func (c *PanelGridContainer) Dragged(ev *fyne.DragEvent) {
	if JM.AppStatusManager.IsDraggable() {
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
		sourceY := c.Position().Y
		scrollStep := float32(10)
		edgeThreshold := float32(30)
		lastDirection := 0

		go func() {
			ticker := time.NewTicker(c.fps)
			defer ticker.Stop()

			for c.dragging {
				<-ticker.C

				targetY := c.dragPosition.Y - c.dragCursorOffset.Y
				direction := 0

				if targetY < sourceY-edgeThreshold {
					direction = -1
				} else if targetY > sourceY+edgeThreshold {
					direction = 1
				}

				// Only scroll if direction is stable and non-zero
				if direction != 0 && direction == lastDirection {
					switch direction {
					case -1:
						JM.AppLayoutManager.ScrollBy(-scrollStep)
					case 1:
						JM.AppLayoutManager.ScrollBy(scrollStep)
					}
				}

				lastDirection = direction
			}
		}()
	}
}

func (c *PanelGridContainer) DragEnd() {
	c.dragging = false
}

func (c *PanelGridContainer) UpdatePanelsContent(shouldUpdate func(pdt *JT.PanelDataType) bool) {
	for _, obj := range c.Objects {
		if panel, ok := obj.(*PanelDisplay); ok {

			pdt := JT.BP.GetDataByID(panel.GetTag())

			if shouldUpdate != nil && !shouldUpdate(pdt) {
				continue
			}

			panel.UpdateContent()
		}
	}
}
func NewPanelGridContainer(
	layout *PanelGridLayout,
	Objects []fyne.CanvasObject,
) *PanelGridContainer {
	c := &PanelGridContainer{
		Objects: Objects,
		layout:  layout,
	}

	for _, obj := range c.Objects {
		if panel, ok := obj.(*PanelDisplay); ok {
			panel.container = c
		}
	}

	c.fps = 16 * time.Millisecond
	if JC.IsMobile {
		c.fps = 32 * time.Millisecond
	}

	c.ExtendBaseWidget(c)

	return c
}
