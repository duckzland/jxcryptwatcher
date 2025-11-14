package core

import (
	"image/color"
	"testing"
)

func TestDynamicFormatFloatToString(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{123.456, "123.456"},
		{1.0, "1"},
		{0.0001, "0.0001"},
	}

	for _, tt := range tests {
		got := DynamicFormatFloatToString(tt.input)
		if got != tt.expected {
			t.Errorf("DynamicFormatFloatToString(%v) = %s; want %s", tt.input, got, tt.expected)
		}
	}
}

func TestSetAlpha(t *testing.T) {
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	alpha := float32(128)
	newColor := SetAlpha(c, alpha)

	r, g, b, a := newColor.RGBA()

	if r != 100*257 || g != 150*257 || b != 200*257 || a != 128*257 {
		t.Errorf("SetAlpha failed: got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}
}

func TestTruncateTextWithEstimation(t *testing.T) {
	str := "This is a long sentence that should be truncated"
	maxWidth := float32(100)
	fontSize := float32(12)

	result := TruncateTextWithEstimation(str, maxWidth, fontSize)
	if len(result) >= len(str) {
		t.Errorf("Expected truncation, got: %s", result)
	}

	short := "Hi"
	result = TruncateTextWithEstimation(short, maxWidth, fontSize)
	if result != short {
		t.Errorf("Expected no truncation, got: %s", result)
	}
}

func TestFormatShortCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$123", "$123.00"},
		{"$1234", "$1.23K"},
		{"$1234567", "$1.23M"},
		{"$1234567890", "$1.23B"},
		{"$1234567890123", "$1.23T"},
		{"not-a-number", "not-a-number"},
	}

	for _, tt := range tests {
		got := FormatShortCurrency(tt.input)
		if got != tt.expected {
			t.Errorf("FormatShortCurrency(%s) = %s; want %s", tt.input, got, tt.expected)
		}
	}
}

func TestExtractLeadingNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123abc", 123},
		{"42", 42},
		{"abc123", -1},
		{STRING_EMPTY, -1},
	}

	for _, tt := range tests {
		got := ExtractLeadingNumber(tt.input)
		if got != tt.expected {
			t.Errorf("ExtractLeadingNumber(%s) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}

func TestSearchableExtractNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123|hello", 123},
		{"42|world", 42},
		{"|missing", -1},
		{"abc|fail", -1},
		{STRING_EMPTY, -1},
	}

	for _, tt := range tests {
		got := SearchableExtractNumber(tt.input)
		if got != tt.expected {
			t.Errorf("SearchableExtractNumber(%s) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}
