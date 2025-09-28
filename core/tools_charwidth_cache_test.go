package core

import (
	"testing"
)

func TestCWCacheInit(t *testing.T) {
	c := &cwCache{}
	c.Init()

	if c.store == nil {
		t.Error("Expected store to be initialized")
	}

	// Check mutex lock/unlock doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Mutex caused panic: %v", r)
			}
		}()
		c.mu.Lock()
		_ = c
		c.mu.Unlock()
	}()
}

func TestCWCacheAddAndGet(t *testing.T) {
	c := &cwCache{}
	c.Init()

	c.Add(42, 3.14)
	val, ok := c.Get(42)

	if !ok {
		t.Error("Expected key 42 to exist")
	}
	if val != 3.14 {
		t.Errorf("Expected value 3.14, got %f", val)
	}
}

func TestCWCacheHas(t *testing.T) {
	c := &cwCache{}
	c.Init()

	c.Add(7, 1.23)
	if !c.Has(7) {
		t.Error("Expected Has(7) to return true")
	}
	if c.Has(999) {
		t.Error("Expected Has(999) to return false")
	}
}

func TestCWCacheKeys(t *testing.T) {
	c := &cwCache{}
	c.Init()

	c.Add(1, 1.0)
	c.Add(2, 2.0)
	c.Add(3, 3.0)

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestRegisterAndUseCharWidthCache(t *testing.T) {
	c1 := RegisterCharWidthCache()
	c2 := UseCharWidthCache()

	if c1 != c2 {
		t.Error("Expected RegisterCharWidthCache and UseCharWidthCache to return the same instance")
	}
}
