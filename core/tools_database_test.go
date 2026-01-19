package core

import (
	"testing"
	"time"
)

func TestDatabase_Reset(t *testing.T) {
	var d Database

	// populate
	d.data.Store("a", 1)
	now := time.Now()
	d.timestamp.Store(&now)

	d.Reset()

	if d.HasData() {
		t.Errorf("expected data to be empty after Reset")
	}

	if d.IsUpdatedAt() != nil {
		t.Errorf("expected timestamp to be nil after Reset")
	}
}

func TestDatabase_SoftReset(t *testing.T) {
	var d Database

	now := time.Now()
	d.timestamp.Store(&now)

	d.SoftReset()

	if d.IsUpdatedAt() != nil {
		t.Errorf("expected timestamp to be nil after SoftReset")
	}
}

func TestDatabase_HasAndHasData(t *testing.T) {
	var d Database

	if d.HasData() {
		t.Errorf("expected empty database initially")
	}

	d.data.Store("x", 123)

	if !d.Has("x") {
		t.Errorf("expected Has(x) to be true")
	}

	if !d.HasData() {
		t.Errorf("expected HasData to be true after storing")
	}

	if d.IsEmpty() {
		t.Errorf("expected IsEmpty to be false after storing")
	}
}

func TestDatabase_UpdatedAt(t *testing.T) {
	var d Database

	if d.IsUpdatedAt() != nil {
		t.Errorf("expected nil timestamp initially")
	}

	now := time.Now()
	d.UpdatedAt(&now)

	got := d.IsUpdatedAt()
	if got == nil || !got.Equal(now) {
		t.Errorf("expected timestamp to be set")
	}
}

func TestDatabase_ShouldRefresh_NoTimestamp(t *testing.T) {
	var d Database
	d.SetUpdateTreshold(10 * time.Second)

	if !d.ShouldRefresh() {
		t.Errorf("expected ShouldRefresh to be true when no timestamp is set")
	}
}

func TestDatabase_ShouldRefresh_NoThreshold(t *testing.T) {
	var d Database

	now := time.Now()
	d.UpdatedAt(&now)

	if !d.ShouldRefresh() {
		t.Errorf("expected ShouldRefresh to be true when no threshold is set")
	}
}

func TestDatabase_ShouldRefresh_True(t *testing.T) {
	var d Database

	past := time.Now().Add(-1 * time.Hour)
	d.UpdatedAt(&past)
	d.SetUpdateTreshold(10 * time.Second)

	if !d.ShouldRefresh() {
		t.Errorf("expected ShouldRefresh to be true when last update is old")
	}
}

func TestDatabase_ShouldRefresh_False(t *testing.T) {
	var d Database

	now := time.Now()
	d.UpdatedAt(&now)
	d.SetUpdateTreshold(1 * time.Hour)

	if d.ShouldRefresh() {
		t.Errorf("expected ShouldRefresh to be false when last update is recent")
	}
}
