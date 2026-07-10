package gatewaymiddleware

import (
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type tokenCacheEntry struct {
	claims jwt.MapClaims
	expiry time.Time
}

type TokenCache struct {
	mu          sync.RWMutex
	store       map[string]tokenCacheEntry
	maxCapacity int
}

func NewTokenCache(maxCapacity int) *TokenCache {
	if maxCapacity <= 0 {
		maxCapacity = 10000 // default capacity
	}
	return &TokenCache{
		store:       make(map[string]tokenCacheEntry),
		maxCapacity: maxCapacity,
	}
}

func (c *TokenCache) Get(token string) (jwt.MapClaims, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[token]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiry) {
		return nil, false
	}

	return entry.claims, true
}

func (c *TokenCache) Set(token string, claims jwt.MapClaims, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Enforce capacity limit to prevent memory exhaustion from random token attacks
	if len(c.store) >= c.maxCapacity {
		// Evict expired entries first to free space
		now := time.Now()
		for k, v := range c.store {
			if now.After(v.expiry) {
				delete(c.store, k)
			}
		}

		// If still at capacity, clear everything to avoid memory leak
		if len(c.store) >= c.maxCapacity {
			c.store = make(map[string]tokenCacheEntry)
		}
	}

	c.store[token] = tokenCacheEntry{
		claims: claims,
		expiry: time.Now().Add(ttl),
	}
}
