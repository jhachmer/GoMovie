// Package cache implements a cache and its necessary internal methods
package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	mutex         sync.RWMutex
	cacheMap      map[K]*CacheItem[V]
	cacheDuration time.Duration
	cleanInterval time.Duration
	cleanTicker   *time.Ticker
	cleanChan     chan bool
	cleanFunc     func(V)
}

type CacheItem[V any] struct {
	value        V
	lastAccessed time.Time
}

// NewCache creates a new Cache
// K is type for keys, they have to be comparable
// V is type for cache items, this can be any type
// cleanInterval sets the time between checking of cache values
// cacheDuration is the maximum duration an item will live in cache when unused
// cleanFunc sets a function which is called, when values get cleaned up (e.g. closing of resources)
func NewCache[K comparable, V any](cleanInterval, cacheDuration time.Duration, cleanFunc func(V)) *Cache[K, V] {
	cache := &Cache[K, V]{
		cacheMap:      make(map[K]*CacheItem[V]),
		cacheDuration: cacheDuration,
		cleanInterval: cleanInterval,
		cleanTicker:   time.NewTicker(cleanInterval),
		cleanChan:     make(chan bool),
		cleanFunc:     cleanFunc,
	}
	go cache.cleanupLoop()
	return cache
}

// cleanupLoop is started in a separate goroutine upon creation of a new cache
// searches for expired cache items in intervals defined upon cache creation
func (c *Cache[K, V]) cleanupLoop() {
	for {
		select {
		case <-c.cleanTicker.C:
			c.mutex.RLock()
			for k, v := range c.cacheMap {
				c.mutex.RUnlock()
				if time.Since(v.lastAccessed) > c.cacheDuration {
					c.mutex.Lock()
					if c.cleanFunc != nil {
						c.cleanFunc(v.value)
					}
					delete(c.cacheMap, k)
					c.mutex.Unlock()
				}
				c.mutex.RLock()
			}
			c.mutex.RUnlock()
		case <-c.cleanChan:
			return
		}
	}
}

// Get retrieves an item from the cache by its key
// returns ok = true if item is found
// returns ok = false if the given key is not present
// updates lastAccessed with the time of access
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	item, ok := c.cacheMap[key]
	defer c.mutex.Unlock()
	if !ok {
		var i V
		return i, false
	}
	item.lastAccessed = time.Now()
	return item.value, true
}

// Set adds value of type V to the Cache map, retrievable with key of type K
// updates value if key is already present in cache and calls cleanupFunc for old value
// if key is not present a new CacheItem is created
func (c *Cache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if item, ok := c.cacheMap[key]; ok {
		if c.cleanFunc != nil {
			c.cleanFunc(item.value)
		}
		item.value = value
		item.lastAccessed = time.Now()
	} else {
		c.cacheMap[key] = &CacheItem[V]{
			value:        value,
			lastAccessed: time.Now(),
		}
	}
}

// Delete deletes CacheItem with key of type K from the cache map
func (c *Cache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.cacheMap, key)
}

// Close closes the Cache
// stops internal ticker, closes cleanup channel to stop cleanup loop goroutine
// calls cleanup function for every item in cache map
func (c *Cache[K, V]) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cleanTicker.Stop()
	close(c.cleanChan)

	for key, item := range c.cacheMap {
		if c.cleanFunc != nil {
			c.cleanFunc(item.value)
		}
		delete(c.cacheMap, key)
	}
}
