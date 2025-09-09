package tickers

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/animations"
	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

type TickerLayout struct{}

func (p *TickerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 4 {
		return
	}
	bg := objects[0]
	title := objects[1]
	content := objects[2]
	status := objects[3]
	spacer := float32(-2)

	bg.Resize(size)
	bg.Move(fyne.NewPos(0, 0))

	centerItems := []fyne.CanvasObject{}
	for _, obj := range []fyne.CanvasObject{title, content, status} {
		if obj.Visible() && obj.MinSize().Height > 0 {
			centerItems = append(centerItems, obj)
		}
	}

	var totalHeight float32
	for _, obj := range centerItems {
		totalHeight += obj.MinSize().Height
	}

	totalHeight += spacer * float32(len(centerItems)-1)

	startY := (size.Height - totalHeight) / 2
	currentY := startY

	for _, obj := range centerItems {
		objSize := obj.MinSize()
		obj.Move(fyne.NewPos((size.Width-objSize.Width)/2, currentY))
		obj.Resize(objSize)
		currentY += objSize.Height + spacer
	}
}

func (p *TickerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)

	for _, obj := range objects[1:4] {
		if obj.Visible() && obj.MinSize().Height > 0 {
			size := obj.MinSize()
			if size.Width > width {
				width = size.Width
			}
			height += size.Height
		}
	}

	return fyne.NewSize(width, height)
}

type TickerDisplay struct {
	widget.BaseWidget
	tag        string
	title      string
	content    fyne.CanvasObject
	child      fyne.CanvasObject
	background *canvas.Rectangle
	refTitle   *canvas.Text
	refContent *canvas.Text
	refStatus  *canvas.Text
}

func NewTickerDisplay(
	tdt *JT.TickerDataType,
) *TickerDisplay {

	uuid := JC.CreateUUID()
	tdt.ID = uuid

	title := canvas.NewText("", JC.TextColor)
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = JC.TickerTitleSize

	status := canvas.NewText("", JC.TextColor)
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}
	status.TextSize = JC.TickerTitleSize

	content := canvas.NewText("", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = JC.TickerContentSize

	background := canvas.NewRectangle(JC.TickerBG)
	background.SetMinSize(fyne.NewSize(100, 100))
	background.CornerRadius = JC.TickerBorderRadius

	str := tdt.GetData()
	ticker := &TickerDisplay{
		tag: uuid,
		content: container.New(&TickerLayout{},
			background,
			title,
			content,
			status,
		),
		title:      tdt.Title,
		background: background,
		refTitle:   title,
		refContent: content,
		refStatus:  status,
	}

	ticker.ExtendBaseWidget(ticker)

	str.AddListener(binding.NewDataListener(func() {

		pkt := JT.BT.GetData(ticker.GetTag())
		if pkt == nil {
			return
		}

		ticker.UpdateContent()

		JA.StartFlashingText(content, 50*time.Millisecond, JC.TextColor, 1)
	}))

	ticker.UpdateContent()

	JA.FadeInBackground(background, 100*time.Millisecond, nil)

	return ticker
}

func (h *TickerDisplay) UpdateContent() {

	pwidth := h.Size().Width
	pkt := JT.BT.GetData(h.GetTag())
	if pkt == nil {
		return
	}

	switch pkt.Status {
	case JC.STATE_ERROR:
		h.refStatus.Text = "Error loading data"
		h.refStatus.Show()
		h.refTitle.Hide()
		h.refContent.Hide()
		h.background.FillColor = JC.ErrorColor

	case JC.STATE_LOADING:
		h.refStatus.Text = "Loading..."
		h.refStatus.Show()
		h.refTitle.Hide()
		h.refContent.Hide()
		h.background.FillColor = JC.PanelBG

	default:
		h.refTitle.Text = JC.TruncateText(h.title, pwidth-20, h.refTitle.TextSize)
		h.refContent.Text = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.refContent.TextSize)
		h.refStatus.Hide()
		h.refTitle.Show()
		h.refContent.Show()

		h.background.FillColor = JC.TickerBG

		if pkt.Type == "altcoin_index" {
			percentage, _ := strconv.ParseInt(pkt.Get(), 10, 64)

			switch {
			case percentage >= 75:
				h.background.FillColor = JC.OrangeColor
			case percentage >= 60:
				h.background.FillColor = JC.LightOrangeColor
			case percentage >= 40:
				h.background.FillColor = JC.LightPurpleColor
			default:
				h.background.FillColor = JC.BlueColor
			}
		}

		if pkt.Type == "feargreed" {
			index, _ := strconv.ParseInt(pkt.Get(), 10, 64)

			switch {
			case index >= 75:
				h.background.FillColor = JC.GreenColor
			case index >= 55:
				h.background.FillColor = JC.TealGreenColor
			case index >= 45:
				h.background.FillColor = JC.YellowColor
			case index >= 25:
				h.background.FillColor = JC.OrangeColor
			default:
				h.background.FillColor = JC.RedColor
			}
		}

		if pkt.Type == "market_cap" {
			raw := JT.TickerCache.Get("market_cap_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)

			if index > 0 {
				h.background.FillColor = JC.GreenColor
			} else if index < 0 {
				h.background.FillColor = JC.RedColor
			}
		}

		if pkt.Type == "cmc100" {
			raw := JT.TickerCache.Get("cmc100_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)

			if index >= 0 {
				h.background.FillColor = JC.GreenColor
			} else if index < 0 {
				h.background.FillColor = JC.RedColor
			}
		}
	}

}

func (h *TickerDisplay) GetTag() string {
	return h.tag
}

func (h *TickerDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *TickerDisplay) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}
