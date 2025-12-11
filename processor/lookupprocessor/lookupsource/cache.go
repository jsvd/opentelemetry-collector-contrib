// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package lookupsource // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor/lookupsource"

import (
	"context"
	"sync"
	"time"
)

type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"`

	Size int `mapstructure:"size"`

	TTL time.Duration `mapstructure:"ttl"`

	// NegativeTTL is the time-to-live for negative cache entries (not found results).
	// Set to 0 to disable negative caching.
	// Default: 0 (disabled)
	NegativeTTL time.Duration `mapstructure:"negative_ttl"`
}

type cacheEntry struct {
	value     any
	found     bool
	expiresAt time.Time
}

func (e *cacheEntry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

type Cache struct {
	config  CacheConfig
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	order   []string
}

func NewCache(cfg CacheConfig) *Cache {
	size := cfg.Size
	if size <= 0 {
		size = 1000
	}
	return &Cache{
		config:  cfg,
		entries: make(map[string]*cacheEntry, size),
		order:   make([]string, 0, size),
	}
}

// retrieves a value from the cache and indicates whether the original lookup found a value.
// Returns (value, lookupFound, cacheHit).
func (c *Cache) get(key string) (any, bool, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false, false
	}

	if entry.isExpired() {
		c.mu.Lock()
		c.removeEntryLocked(key)
		c.mu.Unlock()
		return nil, false, false
	}

	return entry.value, entry.found, true
}

// adds or updates a value in the cache.
func (c *Cache) set(key string, value any, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ttl := c.config.TTL
	if !found {
		if c.config.NegativeTTL == 0 {
			return
		}
		ttl = c.config.NegativeTTL
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if _, exists := c.entries[key]; exists {
		c.entries[key] = &cacheEntry{
			value:     value,
			found:     found,
			expiresAt: expiresAt,
		}
		// MRU
		c.moveToEndLocked(key)
		return
	}

	// Evict oldest entries if at capacity
	for len(c.entries) >= c.config.Size && len(c.order) > 0 {
		oldest := c.order[0]
		delete(c.entries, oldest)
		c.order = c.order[1:]
	}

	c.entries[key] = &cacheEntry{
		value:     value,
		found:     found,
		expiresAt: expiresAt,
	}
	c.order = append(c.order, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
	c.order = make([]string, 0)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// removeEntryLocked removes an entry from the cache.
// Caution: Must be called with the lock held.
func (c *Cache) removeEntryLocked(key string) {
	delete(c.entries, key)
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// moveToEndLocked moves a key to the end of the order slice.
// Caution: Must be called with the lock held.
func (c *Cache) moveToEndLocked(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

// WrapWithCache wraps a lookup function with caching.
//
// The cache supports:
//   - LRU eviction when max size is reached
//   - TTL-based expiration for positive results
//   - Negative caching (caching "not found" results) with separate TTL
//
// Example:
//
//	cache := lookupsource.NewCache(lookupsource.CacheConfig{
//	    Enabled:     true,
//	    Size:        1000,
//	    TTL:         5 * time.Minute,
//	    NegativeTTL: 1 * time.Minute,
//	})
//	cachedLookup := lookupsource.WrapWithCache(cache, myLookupFunc)
func WrapWithCache(cache *Cache, fn LookupFunc) LookupFunc {
	if cache == nil || !cache.config.Enabled {
		return fn
	}
	return func(ctx context.Context, key string) (any, bool, error) {
		if val, lookupFound, cacheHit := cache.get(key); cacheHit {
			return val, lookupFound, nil
		}

		val, found, err := fn(ctx, key)
		if err != nil {
			return nil, false, err
		}

		cache.set(key, val, found)
		return val, found, nil
	}
}
