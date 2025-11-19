package core

import (
	"context"
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
