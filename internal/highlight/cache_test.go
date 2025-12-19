package highlight

import (
	"fmt"
	"sync"
	"testing"
)

func TestCacheBasics(t *testing.T) {
	cache := NewCache(10)

	// Test set and get
	cache.Set("key1", "value1")
	val, found := cache.Get("key1")
	if !found {
		t.Error("cache should find set value")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %q", val)
	}

	// Test miss
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("cache should return false for nonexistent key")
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := NewCache(10)

	// Set initial value
	cache.Set("key1", "value1")

	// Update value
	cache.Set("key1", "value2")

	val, found := cache.Get("key1")
	if !found {
		t.Error("cache should find updated value")
	}
	if val != "value2" {
		t.Errorf("expected 'value2', got %q", val)
	}
}

func TestCacheEviction(t *testing.T) {
	cache := NewCache(3) // Small capacity

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Add one more (should evict oldest)
	cache.Set("key4", "value4")

	// key1 should be evicted (it was the oldest)
	_, found := cache.Get("key1")
	if found {
		t.Error("oldest key should be evicted")
	}

	// Other keys should still exist
	_, found = cache.Get("key4")
	if !found {
		t.Error("newest key should exist")
	}

	_, found = cache.Get("key2")
	if !found {
		t.Error("key2 should still exist")
	}
}

func TestCacheLRU(t *testing.T) {
	cache := NewCache(3)

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Access key1 (makes it most recently used)
	cache.Get("key1")

	// Add new key (should evict key2, not key1)
	cache.Set("key4", "value4")

	// key1 should still exist (recently accessed)
	_, found := cache.Get("key1")
	if !found {
		t.Error("recently accessed key1 should not be evicted")
	}

	// key2 should be evicted (least recently used)
	_, found = cache.Get("key2")
	if found {
		t.Error("key2 should be evicted as LRU")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(10)

	// Add some items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear cache
	cache.Clear()

	// All keys should be gone
	_, found := cache.Get("key1")
	if found {
		t.Error("cache should be empty after Clear")
	}
	_, found = cache.Get("key2")
	if found {
		t.Error("cache should be empty after Clear")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(100)

	// Run concurrent reads and writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			cache.Set(key, value)
			cache.Get(key)
		}(i)
	}
	wg.Wait()

	// No panics = success
	// Check that some values are present
	val, found := cache.Get("key50")
	if !found {
		t.Error("expected key50 to be in cache")
	}
	if val != "value50" {
		t.Errorf("expected 'value50', got %q", val)
	}
}

func TestCacheKey(t *testing.T) {
	// Test that cache key generation is deterministic
	key1 := cacheKey("code", "python")
	key2 := cacheKey("code", "python")
	if key1 != key2 {
		t.Error("same input should generate same cache key")
	}

	// Different inputs should generate different keys
	key3 := cacheKey("code", "go")
	if key1 == key3 {
		t.Error("different language should generate different cache key")
	}

	key4 := cacheKey("different", "python")
	if key1 == key4 {
		t.Error("different code should generate different cache key")
	}
}
