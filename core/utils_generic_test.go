package core

import (
	"runtime"
	"testing"
	"time"
)

func TestTotalCPU(t *testing.T) {
	expected := runtime.NumCPU()
	if TotalCPU() != expected {
		t.Errorf("Expected TotalCPU to be %d, got %d", expected, TotalCPU())
	}
}

func TestMaximumThreads(t *testing.T) {
	tests := []struct {
		capLimit int
		expected int
	}{
		{capLimit: 1000, expected: TotalCPU() / 2},
		{capLimit: 1, expected: 1},
		{capLimit: 0, expected: 1},
		{capLimit: TotalCPU(), expected: TotalCPU() / 2},
	}

	for _, tt := range tests {
		got := MaximumThreads(tt.capLimit)
		if got != tt.expected && !(tt.capLimit < tt.expected && got == tt.capLimit) {
			t.Errorf("MaximumThreads(%d) = %d, want %d", tt.capLimit, got, tt.expected)
		}
	}
}

func TestReorderByMatch(t *testing.T) {
	input := []string{
		"10005|ZOOT - Zoo Token",
		"10011|COW - CoinWind",
		"10021|DINU - Dogey-Inu",
		"1004|HNC - HNC COIN",
		"10030|XMS - Mars Ecosystem Token",
		"10|FRC - Freicoin",
	}

	expected := []string{
		"10|FRC - Freicoin",
		"1004|HNC - HNC COIN",
		"10005|ZOOT - Zoo Token",
		"10011|COW - CoinWind",
		"10021|DINU - Dogey-Inu",
		"10030|XMS - Mars Ecosystem Token",
	}

	result := ReorderByMatch(append([]string{}, input...), "irrelevant")

	if !EqualStringSlices(result, expected) {
		t.Errorf("ReorderByMatch failed.\nGot:      %v\nExpected: %v", result, expected)
	}
}

func TestReorderSearchable(t *testing.T) {
	input := []string{
		"10005|ZOOT - Zoo Token",
		"10011|COW - CoinWind",
		"10021|DINU - Dogey-Inu",
		"1004|HNC - HNC COIN",
		"10030|XMS - Mars Ecosystem Token",
		"10|FRC - Freicoin",
	}

	expected := []string{
		"10|FRC - Freicoin",
		"1004|HNC - HNC COIN",
		"10005|ZOOT - Zoo Token",
		"10011|COW - CoinWind",
		"10021|DINU - Dogey-Inu",
		"10030|XMS - Mars Ecosystem Token",
	}

	result := ReorderSearchable(append([]string{}, input...)) // avoid mutating original

	if !EqualStringSlices(result, expected) {
		t.Errorf("ReorderSearchable failed.\nGot:      %v\nExpected: %v", result, expected)
	}
}

func TestCreateUUID(t *testing.T) {
	id := CreateUUID()
	if len(id) == 0 {
		t.Error("Expected non-empty UUID string")
	}
}

func TestGetMonthBounds(t *testing.T) {
	tm := time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC)
	start, end := GetMonthBounds(tm)

	expectedStart := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedEnd := time.Date(2025, time.September, 30, 23, 59, 59, 0, time.UTC).Unix()

	if start != expectedStart || end != expectedEnd {
		t.Errorf("GetMonthBounds failed. Got (%d, %d), want (%d, %d)", start, end, expectedStart, expectedEnd)
	}
}

func TestEqualStringSlices(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a", "b", "c"}
	c := []string{"a", "b", "d"}

	if !EqualStringSlices(a, b) {
		t.Error("Expected slices to be equal")
	}
	if EqualStringSlices(a, c) {
		t.Error("Expected slices to be unequal")
	}
}

func TestEqualIntSlices(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	c := []int{1, 2, 4}

	if !EqualIntSlices(a, b) {
		t.Error("Expected slices to be equal")
	}
	if EqualIntSlices(a, c) {
		t.Error("Expected slices to be unequal")
	}
}
