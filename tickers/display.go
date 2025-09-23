package tickers

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
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
	title      *tickerText
	content    *tickerText
	status     *tickerText
	state      int
}

func NewtickerDisplay(tdt JT.TickerData) *tickerDisplay {
	uuid := JC.CreateUUID()
	tdt.SetID(uuid)

	tc := JC.ThemeColor(theme.ColorNameForeground)

	tl := &tickerLayout{
		background: canvas.NewRectangle(JC.ThemeColor(JC.ColorNameTickerBG)),
		title:      NewTickerText("", tc, JC.ThemeSize(JC.SizeTickerTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: false}),
		status:     NewTickerText("", tc, JC.ThemeSize(JC.SizeTickerTitle), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		content:    NewTickerText("", tc, JC.ThemeSize(JC.SizeTickerContent), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}

	tl.background.CornerRadius = JC.ThemeSize(JC.SizeTickerBorderRadius)

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

	pkt := JT.UseTickerMaps().GetData(h.GetTag())

	if pkt == nil {
		return
	}

	title := ""
	status := ""
	content := ""
	state := h.state
	background := JC.ThemeColor(JC.ColorNameTickerBG)
	isNewContent := false
	h.state = pkt.GetStatus()

	switch h.state {
	case JC.STATE_ERROR:
		status = "Error loading data"
		background = JC.ThemeColor(JC.ColorNameError)

	case JC.STATE_LOADING:
		status = "Loading..."
		background = JC.ThemeColor(JC.ColorNameTickerBG)

	default:

		title = pkt.GetTitle()
		content = pkt.FormatContent()

		if h.content.GetText().Text != content {
			isNewContent = true
		}

		if pkt.IsType("altcoin_index") {
			percentage, _ := strconv.ParseInt(pkt.Get(), 10, 64)
			switch {
			case percentage >= 75:
				background = JC.ThemeColor(JC.ColorNameBlue)
			case percentage >= 50:
				background = JC.ThemeColor(JC.ColorNameLightPurple)
			case percentage >= 25:
				background = JC.ThemeColor(JC.ColorNameLightOrange)
			default:
				background = JC.ThemeColor(JC.ColorNameOrange)
			}
		}

		if pkt.IsType("feargreed") {
			index, _ := strconv.ParseInt(pkt.Get(), 10, 64)
			switch {
			case index >= 75:
				background = JC.ThemeColor(JC.ColorNameGreen)
			case index >= 55:
				background = JC.ThemeColor(JC.ColorNameTeal)
			case index >= 45:
				background = JC.ThemeColor(JC.ColorNameYellow)
			case index >= 25:
				background = JC.ThemeColor(JC.ColorNameOrange)
			default:
				background = JC.ThemeColor(JC.ColorNameRed)
			}
		}

		if pkt.IsType("market_cap") {
			raw := JT.UseTickerCache().Get("market_cap_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)
			if index > 0 {
				background = JC.ThemeColor(JC.ColorNameGreen)
			} else if index < 0 {
				background = JC.ThemeColor(JC.ColorNameRed)
			}
		}

		if pkt.IsType("cmc100") {
			raw := JT.UseTickerCache().Get("cmc100_24_percentage")
			index, _ := strconv.ParseFloat(raw, 64)
			if index >= 0 {
				background = JC.ThemeColor(JC.ColorNameGreen)
			} else if index < 0 {
				background = JC.ThemeColor(JC.ColorNameRed)
			}
		}
	}

	h.title.SetText(title)
	h.status.SetText(status)
	h.content.SetText(content)

	if isNewContent {
		JA.StartFlashingText(h.content.GetText(), 50*time.Millisecond, h.content.GetText().Color, 1)
	}

	if h.background.FillColor != background {
		h.background.FillColor = background
		canvas.Refresh(h.background)
	}

	if h.state != state {
		h.Refresh()
	}
}
