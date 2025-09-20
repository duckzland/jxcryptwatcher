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
	title      string
	content    fyne.CanvasObject
	background *canvas.Rectangle
	refTitle   *canvas.Text
	refContent *canvas.Text
	refStatus  *canvas.Text
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

	tl.background.SetMinSize(fyne.NewSize(100, 100))
	tl.background.CornerRadius = JC.TickerBorderRadius

	str := tdt.GetData()
	ticker := &tickerDisplay{
		tag: uuid,
		content: container.New(
			tl,
			tl.background,
			tl.title,
			tl.content,
			tl.status,
		),
		title:      tdt.GetTitle(),
		background: tl.background,
		refTitle:   tl.title,
		refContent: tl.content,
		refStatus:  tl.status,
	}

	ticker.ExtendBaseWidget(ticker)

	str.AddListener(binding.NewDataListener(func() {
		ticker.UpdateContent()
		JA.StartFlashingText(ticker.refContent, 50*time.Millisecond, JC.TextColor, 1)
	}))

	ticker.UpdateContent()
	JA.FadeInBackground(ticker.background, 100*time.Millisecond, nil)

	return ticker
}

func (h *tickerDisplay) UpdateContent() {
	pwidth := h.Size().Width
	pkt := JT.BT.GetData(h.GetTag())
	if pkt == nil {
		return
	}

	switch pkt.GetStatus() {
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
}

func (h *tickerDisplay) GetTag() string {
	return h.tag
}

func (h *tickerDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *tickerDisplay) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}
