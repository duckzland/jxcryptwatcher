package core

import (
	"math/big"
	"testing"
)

func TestNumDecPlaces(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{123.456, 3},
		{1.0, 0},
		{0.0001, 4},
		{100.123456789, 9},
		{42, 0},
	}

	for _, tt := range tests {
		got := NumDecPlaces(tt.input)
		if got != tt.expected {
			t.Errorf("NumDecPlaces(%v) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}

func TestBigFloatNumDecPlaces(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123.456000", 3},
		{"1.0", 0},
		{"0.0001000", 4},
		{"100.123456789000", 9},
		{"42", 0},
	}

	for _, tt := range tests {
		bf, ok := new(big.Float).SetPrec(256).SetString(tt.input)
		if !ok {
			t.Fatalf("Failed to parse big.Float from %s", tt.input)
		}
		got := BigFloatNumDecPlaces(bf)
		if got != tt.expected {
			t.Errorf("BigFloatNumDecPlaces(%s) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}

func TestToBigFloat(t *testing.T) {
	val := 123.456
	bf := ToBigFloat(val)
	if bf == nil {
		t.Error("ToBigFloat returned nil")
	}
	expected := new(big.Float).SetPrec(256).SetFloat64(val)
	if bf.Cmp(expected) != 0 {
		t.Errorf("ToBigFloat(%v) = %v; want %v", val, bf, expected)
	}
}

func TestToBigString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"123.456", "123.456", true},
		{"0.0001", "0.0001", true},
		{"abc", JC.STRING_EMPTY, false},
	}

	for _, tt := range tests {
		bf, ok := ToBigString(tt.input)
		if ok != tt.valid {
			t.Errorf("ToBigString(%s) validity = %v; want %v", tt.input, ok, tt.valid)
		}
		if ok && bf.Text('f', -1) != tt.expected {
			t.Errorf("ToBigString(%s) = %s; want %s", tt.input, bf.Text('f', -1), tt.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"-42", true},
		{"3.14", false},
		{"abc", false},
	}

	for _, tt := range tests {
		got := IsNumeric(tt.input)
		if got != tt.expected {
			t.Errorf("IsNumeric(%s) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFormatNumberWithCommas(t *testing.T) {
	tests := []struct {
		input    float64
		frac     int
		expected string
	}{
		{1000000.0, 3, "1,000,000"},
		{1234567.89, 2, "1,234,567.89"},
		{0.000001, 6, "0.000001"},
		{987654321.123456, 6, "987,654,321.123456"},
		{42.0, 0, "42"},
		{1000.5, 1, "1,000.5"},
		{1000.0, 0, "1,000"},
	}

	for _, tt := range tests {
		result := FormatNumberWithCommas(tt.input, tt.frac)
		if result != tt.expected {
			t.Errorf("FormatNumberWithCommas(%f, %d) = %s; want %s", tt.input, tt.frac, result, tt.expected)
		}
	}
}
