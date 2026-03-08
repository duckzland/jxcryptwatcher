package types

import (
	"testing"
)

func TestWatcherKeyType_Basic(t *testing.T) {
	w := NewWatcherKey()

	// Generate a key from args
	key := w.GenerateKeyFromArgs(5, 1, 1234.5678, 10, 30, 999999)
	if key == "" {
		t.Errorf("expected non-empty key")
	}

	// Check getters
	if w.GetSent() != 5 {
		t.Errorf("expected Sent=5, got %d", w.GetSent())
	}
	if w.GetOperator() != 1 {
		t.Errorf("expected Operator=1, got %d", w.GetOperator())
	}
	if w.GetRate() != 1234.5678 {
		t.Errorf("expected Rate=1234.5678, got %f", w.GetRate())
	}
	if w.GetLimit() != 10 {
		t.Errorf("expected Limit=10, got %d", w.GetLimit())
	}
	if w.GetDuration() != 30 {
		t.Errorf("expected Duration=30, got %d", w.GetDuration())
	}
	if w.GetTimestamp() != 999999 {
		t.Errorf("expected Timestamp=999999, got %d", w.GetTimestamp())
	}

	w = NewWatcherKey()

	// Generate a key from args
	pdt := panelType{}
	pdt.Sent = 5
	pdt.Operator = 1
	pdt.Rate = 1234.5678
	pdt.Limit = 10
	pdt.Duration = 30
	pdt.Timestamp = 999999

	key = w.GenerateKeyFromPanel(pdt)

	if key == "" {
		t.Errorf("expected from panel non-empty key")
	}

	// Check getters
	if w.GetSent() != 5 {
		t.Errorf("expected from panel Sent=5, got %d", w.GetSent())
	}
	if w.GetOperator() != 1 {
		t.Errorf("expected from panel Operator=1, got %d", w.GetOperator())
	}
	if w.GetRate() != 1234.5678 {
		t.Errorf("expected from panel Rate=1234.5678, got %f", w.GetRate())
	}
	if w.GetLimit() != 10 {
		t.Errorf("expected from panel Limit=10, got %d", w.GetLimit())
	}
	if w.GetDuration() != 30 {
		t.Errorf("expected from panel Duration=30, got %d", w.GetDuration())
	}
	if w.GetTimestamp() != 999999 {
		t.Errorf("expected from panel Timestamp=999999, got %d", w.GetTimestamp())
	}

	if w.GetRawValue() != "5|1|1234.5678|10|30|999999" {
		t.Errorf("expected raw: %s, got %s", "5|1|1234.5678|10|30|999999", w.GetRawValue())
	}

	npt := w.ToPanel(panelType{})

	if npt.Sent != 5 {
		t.Errorf("expected from ToPanel Sent=5, got %d", npt.Sent)
	}
	if npt.Operator != 1 {
		t.Errorf("expected from ToPanel Operator=1, got %d", npt.Operator)
	}
	if npt.Rate != 1234.5678 {
		t.Errorf("expected from ToPanel Rate=1234.5678, got %f", npt.Rate)
	}
	if npt.Limit != 10 {
		t.Errorf("expected from ToPanel Limit=10, got %d", npt.Limit)
	}
	if npt.Duration != 30 {
		t.Errorf("expected from ToPanel Duration=30, got %d", npt.Duration)
	}
	if npt.Timestamp != 999999 {
		t.Errorf("expected from ToPanel Timestamp=999999, got %d", npt.Timestamp)
	}

	// Update sent and timestamp
	w.UpdateSent(42)
	if w.GetSent() != 42 {
		t.Errorf("expected Sent=42 after update, got %d", w.GetSent())
	}
	w.UpdateTimestamp(1234567890)
	if w.GetTimestamp() != 1234567890 {
		t.Errorf("expected Timestamp=1234567890 after update, got %d", w.GetTimestamp())
	}
}

func TestWatcherKeyType_FormattedRate(t *testing.T) {
	w := NewWatcherKey()
	// small rate to trigger frac=4
	w.GenerateKeyFromArgs(0, 0, 0.012345, 0, 0, 0)

	formatted := w.GetFormattedRateString()
	if formatted == "" {
		t.Errorf("expected formatted string, got empty")
	}
	t.Logf("Formatted rate: %s", formatted)

	// larger rate should use 2 decimals minimum
	w.GenerateKeyFromArgs(0, 0, 12345.6789, 0, 0, 0)
	formatted2 := w.GetFormattedRateString()
	t.Logf("Formatted rate: %s", formatted2)
}
