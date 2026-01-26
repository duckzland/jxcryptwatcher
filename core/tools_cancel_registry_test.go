package core

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestCancelRegistry(t *testing.T) {
	reg := NewCancelRegistry(10)

	tag := "test:fade"
	_, cancel := context.WithCancel(context.Background())

	// Test Set and Exists
	reg.Set(tag, cancel)
	if !reg.Exists(tag) {
		t.Errorf("Expected tag %q to exist", tag)
	}

	// Test Get
	got, ok := reg.Get(tag)
	if !ok {
		t.Errorf("Expected to retrieve cancel func for tag %q", tag)
	}
	if got == nil {
		t.Errorf("CancelFunc for tag %q should not be nil", tag)
	}

	// Test Delete
	reg.Delete(tag)
	if reg.Exists(tag) {
		t.Errorf("Expected tag %q to be deleted", tag)
	}

	// Test Get after Delete
	_, ok = reg.Get(tag)
	if ok {
		t.Errorf("Expected no cancel func for deleted tag %q", tag)
	}
}

func TestCancelRegistryRange(t *testing.T) {
	reg := NewCancelRegistry(10)

	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		_, cancel := context.WithCancel(context.Background())
		reg.Set(k, cancel)
	}

	seen := make(map[string]bool)
	reg.Range(func(key string, cancel context.CancelFunc) bool {
		seen[key] = true
		return true
	})

	for _, k := range keys {
		if !seen[k] {
			t.Errorf("Range did not visit key %q", k)
		}
	}
}

func TestCancelRegistryDestroy(t *testing.T) {
	reg := NewCancelRegistry(10)

	var canceledCount atomic.Int64

	// Create cancel funcs that increment a counter when called
	makeCancel := func() context.CancelFunc {
		return func() {
			canceledCount.Add(1)
		}
	}

	reg.Set("x", makeCancel())
	reg.Set("y", makeCancel())
	reg.Set("z", makeCancel())

	if reg.Len() != 3 {
		t.Errorf("Expected Len() = 3 before destroy, got %d", reg.Len())
	}

	reg.Destroy()

	// All cancel funcs should have been called
	if canceledCount.Load() != 3 {
		t.Errorf("Expected 3 cancel funcs to be called, got %d", canceledCount.Load())
	}

	// Registry should be empty
	if reg.Len() != 0 {
		t.Errorf("Expected Len() = 0 after destroy, got %d", reg.Len())
	}

	// Range should visit nothing
	visited := false
	reg.Range(func(key string, cancel context.CancelFunc) bool {
		visited = true
		return true
	})
	if visited {
		t.Errorf("Expected no entries in registry after destroy")
	}
}
