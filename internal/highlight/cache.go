package highlight

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Cache provides thread-safe LRU caching for highlighted code
type Cache struct {
	mu       sync.RWMutex
	items    map[string]*cacheEntry
	lru      *list.List
	capacity int
}

type cacheEntry struct {
	key   string
	value string
	elem  *list.Element
}

// NewCache creates a new LRU cache with the specified capacity
func NewCache(capacity int) *Cache {
	return &Cache{
		items:    make(map[string]*cacheEntry, capacity),
		lru:      list.New(),
		capacity: capacity,
	}
}

// Get retrieves a value from the cache and marks it as recently used
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, found := c.items[key]; found {
		// Move to front (most recently used)
		c.lru.MoveToFront(entry.elem)
		return entry.value, true
	}
	return "", false
}

// Set adds or updates a value in the cache
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If exists, update and move to front
	if entry, found := c.items[key]; found {
		entry.value = value
		c.lru.MoveToFront(entry.elem)
		return
	}

	// Evict if at capacity
	if len(c.items) >= c.capacity {
		c.evictOldest()
	}

	// Add new entry
	entry := &cacheEntry{
		key:   key,
		value: value,
	}
	entry.elem = c.lru.PushFront(entry)
	c.items[key] = entry
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheEntry, c.capacity)
	c.lru.Init()
}

// evictOldest removes the least recently used item (caller must hold lock)
func (c *Cache) evictOldest() {
	if c.lru.Len() == 0 {
		return
	}

	oldest := c.lru.Back()
	if oldest != nil {
		c.lru.Remove(oldest)
		entry, ok := oldest.Value.(*cacheEntry)
		if ok {
			delete(c.items, entry.key)
		}
	}
}

// cacheKey generates a cache key from code and language
func cacheKey(code, language string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(code))
	_, _ = h.Write([]byte(language))
	return hex.EncodeToString(h.Sum(nil))
}
