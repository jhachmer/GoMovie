package cache

import (
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	cache := NewCache[string, int](time.Minute, time.Minute, nil)
	cache.Set("key1", 1)

	value, ok := cache.Get("key1")
	if !ok || value != 1 {
		t.Errorf("expected to get 1, got %v", value)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache[string, int](time.Minute, time.Minute, nil)

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Errorf("expected to get false for non-existent key")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache[string, int](time.Minute, time.Minute, nil)
	cache.Set("key1", 1)
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Errorf("expected to get false after deletion")
	}
}

func TestCache_Cleanup(t *testing.T) {
	cleanupCalled := false
	cleanupFunc := func(value int) {
		cleanupCalled = true
	}

	cache := NewCache[string, int](time.Millisecond, time.Millisecond, cleanupFunc)
	cache.Set("key1", 1)

	time.Sleep(10 * time.Millisecond)
	_, ok := cache.Get("key1")
	if ok {
		t.Errorf("expected to get false after cleanup")
	}

	if !cleanupCalled {
		t.Errorf("expected cleanup function to be called")
	}
}

func TestCache_Close(t *testing.T) {
	cleanupCalled := false
	cleanupFunc := func(value int) {
		cleanupCalled = true
	}

	cache := NewCache[string, int](time.Minute, time.Minute, cleanupFunc)
	cache.Set("key1", 1)
	cache.Close()

	_, ok := cache.Get("key1")
	if ok {
		t.Errorf("expected to get false after close")
	}

	if !cleanupCalled {
		t.Errorf("expected cleanup function to be called on close")
	}
}
