package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type navigableList struct {
	widget.List
	selected        int
	setTextFromMenu func(string)
	hide            func()
	filteredData    []string
	navigating      bool
	visibleCount    int
}

func NewNavigableList(
	setTextFromMenu func(string),
	hide func(),
) *navigableList {

	n := &navigableList{
		selected:        -1,
		setTextFromMenu: setTextFromMenu,
		hide:            hide,
	}

	n.visibleCount = 10

	if JC.IsMobile {
		n.visibleCount = 6
	}

	n.List = widget.List{
		Length: func() int {
			if n.visibleCount > len(n.filteredData) {
				return len(n.filteredData)
			}
			return n.visibleCount
		},
		CreateItem: func() fyne.CanvasObject {
			return NewSelectableText()
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= n.visibleCount-10 && n.visibleCount < len(n.filteredData) {
				n.visibleCount += 50
				if n.visibleCount > len(n.filteredData) {
					n.visibleCount = len(n.filteredData)
				}
				n.Refresh()
			}

			item := n.filteredData[i]
			if st, ok := o.(*SelectableText); ok {
				st.SetText(item)
				st.SetIndex(i)
				st.SetParent(n)
			}
		},
		OnSelected: func(i widget.ListItemID) {
			if !n.navigating && i > -1 {
				item := n.filteredData[i]
				n.setTextFromMenu(item)
			}
			n.navigating = false
		},
	}

	n.ExtendBaseWidget(n)

	return n
}

func (n *navigableList) FocusGained() {
}

func (n *navigableList) FocusLost() {
}

func (n *navigableList) SetFilteredData(items []string) {
	if JC.EqualStringSlices(n.filteredData, items) {
		return
	}

	n.Unselect(n.selected)
	n.filteredData = items
	n.selected = -1

	n.visibleCount = 10
	if JC.IsMobile {
		n.visibleCount = 6
	}

	n.List.ScrollToTop()

	for i := 0; i < n.visibleCount && i < len(n.filteredData); i++ {
		n.List.RefreshItem(i)
	}
}

func (n *navigableList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.filteredData)-1 {
			n.selected++
		} else {
			n.selected = 0
		}
		n.navigating = true
		n.Select(n.selected)

	case fyne.KeyUp:
		if n.selected > 0 {
			n.selected--
		} else {
			n.selected = len(n.filteredData) - 1
		}
		n.navigating = true
		n.Select(n.selected)
	case fyne.KeyEscape:
		n.hide()

	}
}

func (n *navigableList) TypedRune(r rune) {}
