package types

import (
	"strconv"
	"strings"

	JC "jxwatcher/core"
)

type watcherKeyType struct {
	value string // Format: "sent|comparator|rate|limit|duration|timestamp"
}

func (p *watcherKeyType) Set(value string) {
	p.value = value
}

func (p *watcherKeyType) IsEmpty() bool {
	return strings.TrimSpace(p.value) == ""
}

func (p *watcherKeyType) UpdateSent(sent int) string {
	pkk := strings.Split(p.value, JC.STRING_PIPE)
	pkk[0] = strconv.Itoa(sent)
	p.value = strings.Join(pkk, JC.STRING_PIPE)
	return p.value
}

func (p *watcherKeyType) UpdateTimestamp(ts int64) string {

	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 6 {
		parts[5] = strconv.FormatInt(ts, 10)
		p.value = strings.Join(parts, JC.STRING_PIPE)
	}
	return p.value
}

func (p *watcherKeyType) GenerateKeyFromPanel(panel panelType) string {

	var b strings.Builder

	b.WriteString(strconv.Itoa(panel.Sent))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(panel.Operator))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.FormatFloat(panel.Rate, 'g', -1, 64))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(panel.Limit))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(panel.Duration))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(panel.Timestamp))

	p.value = b.String()
	return p.value
}

func (p *watcherKeyType) GenerateKeyFromArgs(sent int, operator int, rate float64, limit int, duration int, timestamp int) string {
	var b strings.Builder

	b.WriteString(strconv.Itoa(sent))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(operator))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.FormatFloat(rate, 'g', -1, 64))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(limit))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(duration))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(strconv.Itoa(timestamp))

	p.value = b.String()
	return p.value
}

func (p *watcherKeyType) GetRawValue() string {
	return p.value
}

func (p *watcherKeyType) ToPanel(base panelType) panelType {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) < 6 {
		return base
	}

	if v, err := strconv.Atoi(parts[0]); err == nil {
		base.Sent = v
	}
	if v, err := strconv.Atoi(parts[1]); err == nil {
		base.Operator = v
	}
	if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
		base.Rate = v
	}
	if v, err := strconv.Atoi(parts[3]); err == nil {
		base.Limit = v
	}
	if v, err := strconv.Atoi(parts[4]); err == nil {
		base.Duration = v
	}
	if v, err := strconv.Atoi(parts[5]); err == nil {
		base.Timestamp = v
	}

	return base
}

func (p *watcherKeyType) GetSent() int {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 1 {
		if v, err := strconv.Atoi(parts[0]); err == nil {
			return v
		}
	}
	return -9999
}

func (p *watcherKeyType) GetOperator() int {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 2 {
		if v, err := strconv.Atoi(parts[1]); err == nil {
			return v
		}
	}
	return 0
}

func (p *watcherKeyType) GetRate() float64 {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 3 {
		if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
			return v
		}
	}
	return 1.0
}

func (p *watcherKeyType) GetLimit() int {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 4 {
		if v, err := strconv.Atoi(parts[3]); err == nil {
			return v
		}
	}
	return 3
}

func (p *watcherKeyType) GetDuration() int {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 5 {
		if v, err := strconv.Atoi(parts[4]); err == nil {
			return v
		}
	}
	return 30
}

func (p *watcherKeyType) GetTimestamp() int {
	parts := strings.Split(p.value, JC.STRING_PIPE)
	if len(parts) >= 6 {
		if v, err := strconv.Atoi(parts[5]); err == nil {
			return v
		}
	}
	return 0
}

func (p *watcherKeyType) GetFormattedRateString() string {
	rate := p.GetRate()
	frac := JC.NumDecPlaces(rate)

	if frac < 3 {
		frac = 2
	}

	if rate < 1 {
		frac = 4
	}

	return JC.FormatNumberWithCommas(rate, frac)
}

func NewWatcherKey() *watcherKeyType {
	return &watcherKeyType{}
}
