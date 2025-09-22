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

type tickerDisplay struct {
	widget.BaseWidget
	tag        string
	container  fyne.CanvasObject
	background *canvas.Rectangle
	title      *canvas.Text
	content    *canvas.Text
	status     *canvas.Text
}

func NewtickerDisplay(tdt *JT.TickerDataType) *tickerDisplay {
	uuid := JC.CreateUUID()
	tdt.SetID(uuid)

	tl := &tickerLayout{
		background: canvas.NewRectangle(JC.TickerBG),
		title:      canvas.NewText("", JC.TextColor),
		content:    canvas.NewText("", JC.TextColor),
		status:     canvas.NewText("", JC.TextColor),
	}

	tl.title.Alignment = fyne.TextAlignCenter
	tl.title.TextSize = JC.TickerTitleSize

	tl.status.Alignment = fyne.TextAlignCenter
	tl.status.TextStyle = fyne.TextStyle{Bold: true}
	tl.status.TextSize = JC.TickerTitleSize

	tl.content.Alignment = fyne.TextAlignCenter
	tl.content.TextStyle = fyne.TextStyle{Bold: true}
	tl.content.TextSize = JC.TickerContentSize

	tl.background.CornerRadius = JC.TickerBorderRadius

	str := tdt.GetData()
	ticker := &tickerDisplay{
		tag: uuid,
		container: container.New(
			tl,
			tl.background,
			tl.title,
			tl.content,
			tl.status,
		),
		background: tl.background,
		title:      tl.title,
		content:    tl.content,
		status:     tl.status,
	}

	ticker.ExtendBaseWidget(ticker)

	str.AddListener(binding.NewDataListener(ticker.updateContent))

	ticker.updateContent()
	JA.FadeInBackground(ticker.background, 100*time.Millisecond, nil)

	return ticker
}

func (h *tickerDisplay) GetTag() string {
	return h.tag
}

func (h *tickerDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.container)
}

func (h *tickerDisplay) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

func (h *tickerDisplay) updateContent() {
	pwidth := h.Size().Width
	pkt := JT.BT.GetData(h.GetTag())

	if pkt == nil {
		return
	}

	switch pkt.GetStatus() {
	case JC.STATE_ERROR:
		h.status.Text = "Error loading data"
		h.status.Show()
		h.title.Hide()
		h.content.Hide()
		h.background.FillColor = JC.ErrorColor

	case JC.STATE_LOADING:
		h.status.Text = "Loading..."
		h.status.Show()
		h.title.Hide()
		h.content.Hide()
		h.background.FillColor = JC.PanelBG

	default:
		h.title.Text = JC.TruncateText(pkt.GetTitle(), pwidth-20, h.title.TextSize)
		h.content.Text = JC.TruncateText(pkt.FormatContent(), pwidth-20, h.content.TextSize)
		h.status.Hide()
		h.title.Show()
		h.content.Show()
		h.background.FillColor = JC.TickerBG

		if pkt.IsType("altcoin_index") {
			percentage, _ := strconv.ParseInt(pkt.Get(), 10, 64)
			switch {
			case percentage >= 75:
				h.background.FillColor = JC.BlueColor
			case percentage >= 50:
				h.background.FillColor = JC.LightPurpleColor
			case percentage >= 25:
				h.background.FillColor = JC.LightOrangeColor
			default:
				h.background.FillColor = JC.OrangeColor
			}
		}

		if pkt.IsType("feargreed") {
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

		if pkt.IsType("market_cap") {
			raw := JT.TickerCache.Get("market_cap_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)
			if index > 0 {
				h.background.FillColor = JC.GreenColor
			} else if index < 0 {
				h.background.FillColor = JC.RedColor
			}
		}

		if pkt.IsType("cmc100") {
			raw := JT.TickerCache.Get("cmc100_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)
			if index >= 0 {
				h.background.FillColor = JC.GreenColor
			} else if index < 0 {
				h.background.FillColor = JC.RedColor
			}
		}
	}

	JA.StartFlashingText(h.content, 50*time.Millisecond, JC.TextColor, 1)
	h.Refresh()
}
