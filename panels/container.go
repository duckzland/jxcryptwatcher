package panels

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type panelContainer struct {
	widget.BaseWidget
	Objects          []fyne.CanvasObject
	layout           *panelGridLayout
	position         fyne.Position
	dragging         bool
	dragCursorOffset fyne.Position
	fps              time.Duration
	activeAction     *panelDisplay
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
	if JA.UseStatus().IsDraggable() {
		return
	}

	// Crash fix nil pointer
	if c == nil || ev == nil {
		return
	}

	if JC.IsMobile {
		JA.UseLayout().UseScroll().Dragged(ev)
		return
	}

	c.position = ev.Position

	// Smoother dragging compares to just pass this to scroll directly
	if !c.dragging {
		c.dragging = true

		go func() {
			ticker := time.NewTicker(c.fps)
			defer ticker.Stop()

			lastY := ev.Position.Y

			for c.dragging {
				<-ticker.C

				currentY := c.position.Y
				deltaY := currentY - lastY

				scrollEvent := &fyne.ScrollEvent{
					Scrolled: fyne.Delta{
						DX: 0,
						DY: -deltaY,
					},
				}

				fyne.Do(func() {
					JA.UseLayout().UseScroll().Scrolled(scrollEvent)
				})

				lastY = currentY
			}
		}()
	}
}

func (c *panelContainer) DragEnd() {
	if JC.IsMobile {
		JA.UseLayout().UseScroll().DragEnd()
		return
	}

	c.dragging = false
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

func (c *panelContainer) SetActiveAction(action *panelDisplay) {
	c.activeAction = action
}

func (c *panelContainer) GetActiveAction() *panelDisplay {
	return c.activeAction
}

func (c *panelContainer) HasActiveAction() bool {
	return c.activeAction != nil
}

func (c *panelContainer) ResetActiveAction() {
	c.activeAction = nil
}

func NewPanelContainer(
	layout *panelGridLayout,
	Objects []fyne.CanvasObject,
) *panelContainer {
	c := &panelContainer{
		Objects: Objects,
		layout:  layout,
	}

	c.fps = 16 * time.Millisecond
	if JC.IsMobile {
		c.fps = 32 * time.Millisecond
	}

	c.ExtendBaseWidget(c)

	return c
}
