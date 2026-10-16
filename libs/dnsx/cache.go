package dnsx

import (
	"sync"
	"sync/atomic"
	"time"
)

// cacheEntry holds a cached DNS answer together with its absolute expiry.
type cacheEntry struct {
	data    *ResponseData
	expires time.Time
}

// TTLCache is a concurrency-safe, TTL-bounded cache of DNS answers. Entries
// are automatically considered expired once their TTL elapses; expired
// entries are lazily evicted on Get. It never returns a response beyond its
// validity, and it never attributes one resolver's answer to another (keys
// are host+type+resolver).
type TTLCache struct {
	mu     sync.RWMutex
	items  map[string]cacheEntry
	hits   atomic.Int64
	misses atomic.Int64
}

// NewTTLCache creates an empty TTL cache.
func NewTTLCache() *TTLCache {
	return &TTLCache{items: make(map[string]cacheEntry)}
}

// Get returns the cached answer for key if present and not expired. The
// second return value is false when the entry is missing or expired (expired
// entries are deleted on read).
func (c *TTLCache) Get(key string) (*ResponseData, bool) {
	c.mu.Lock()
	entry, ok := c.items[key]
	if ok && time.Now().Before(entry.expires) {
		c.hits.Add(1)
		c.mu.Unlock()
		return entry.data, true
	}
	if ok {
		delete(c.items, key)
	}
	c.mu.Unlock()
	c.misses.Add(1)
	return nil, false
}

// Set stores data in the cache, expiring after ttl. A ttl <= 0 means the
// entry is stored but considered immediately expired (useful for negative
// caching where the caller decides the TTL).
func (c *TTLCache) Set(key string, data *ResponseData, ttl time.Duration) {
	expires := time.Now().Add(ttl)
	if ttl <= 0 {
		expires = time.Now()
	}
	c.mu.Lock()
	c.items[key] = cacheEntry{data: data, expires: expires}
	c.mu.Unlock()
}

// Delete removes the entry for key if present.
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Len returns the number of entries (including expired ones not yet evicted).
func (c *TTLCache) Len() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

// Stats returns hit/miss counts.
func (c *TTLCache) Stats() (hits, misses int64) {
	return c.hits.Load(), c.misses.Load()
}

// cacheKey builds a deterministic key for a host + question type + resolver,
// so that one resolver's answer is never attributed to another.
func cacheKey(host string, qtype uint16, resolver string) string {
	return host + "/" + dnsTypeToString(qtype) + "/" + resolver
}

// dnsTypeToString returns a short label for a DNS question type.
func dnsTypeToString(qtype uint16) string {
	switch qtype {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 12:
		return "PTR"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 33:
		return "SRV"
	case 48:
		return "DNSKEY"
	case 49:
		return "SPF"
	case 255:
		return "ANY"
	case 257:
		return "CAA"
	case 28:
		return "AAAA"
	case 37:
		return "NSEC"
	case 39:
		return "DNAME"
	case 50:
		return "NSEC3"
	case 51:
		return "NSEC3PARAM"
	default:
		return "TYPE" + string(rune(qtype))
	}
}
