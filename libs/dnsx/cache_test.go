package dnsx

import (
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/projectdiscovery/retryabledns"
	"github.com/stretchr/testify/require"
)

func TestTTLCache_SetGet(t *testing.T) {
	c := NewTTLCache()
	rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com", A: []string{"1.1.1.1"}}}
	c.Set("example.com/A", rd, time.Minute)
	got, ok := c.Get("example.com/A")
	require.True(t, ok)
	require.Equal(t, "1.1.1.1", got.A[0])
}

func TestTTLCache_ExpiresAfterTTL(t *testing.T) {
	c := NewTTLCache()
	rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com", A: []string{"1.1.1.1"}}}
	c.Set("example.com/A", rd, 10*time.Millisecond)
	got, ok := c.Get("example.com/A")
	require.True(t, ok)
	require.Equal(t, "1.1.1.1", got.A[0])

	time.Sleep(30 * time.Millisecond)
	got, ok = c.Get("example.com/A")
	require.False(t, ok)
	require.Nil(t, got)
}

func TestTTLCache_TTLZeroExpiresImmediately(t *testing.T) {
	c := NewTTLCache()
	rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com", A: []string{"1.1.1.1"}}}
	c.Set("example.com/A", rd, 0)
	got, ok := c.Get("example.com/A")
	require.False(t, ok)
	require.Nil(t, got)
}

func TestTTLCache_Delete(t *testing.T) {
	c := NewTTLCache()
	rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com", A: []string{"1.1.1.1"}}}
	c.Set("example.com/A", rd, time.Minute)
	require.Equal(t, 1, c.Len())
	c.Delete("example.com/A")
	require.Equal(t, 0, c.Len())
	_, ok := c.Get("example.com/A")
	require.False(t, ok)
}

func TestTTLCache_Concurrent(t *testing.T) {
	c := NewTTLCache()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com", A: []string{string(rune('0' + i%10))}}}
			c.Set("key", rd, time.Minute)
			_, _ = c.Get("key")
		}(i)
	}
	wg.Wait()
	require.GreaterOrEqual(t, c.Len(), 0)
}

func TestTTLCache_Stats(t *testing.T) {
	c := NewTTLCache()
	rd := &ResponseData{DNSData: &retryabledns.DNSData{Host: "example.com"}}
	c.Set("a", rd, time.Minute)
	c.Get("a") // hit
	c.Get("b") // miss
	h, m := c.Stats()
	require.Equal(t, int64(1), h)
	require.Equal(t, int64(1), m)
}

func TestCacheKey_IncludesResolver(t *testing.T) {
	k1 := cacheKey("example.com", dns.TypeA, "resolver1")
	k2 := cacheKey("example.com", dns.TypeA, "resolver2")
	require.NotEqual(t, k1, k2, "different resolvers must not share a cache key")
	k3 := cacheKey("example.com", dns.TypeA, "resolver1")
	require.Equal(t, k1, k3, "same inputs must produce same key")
}

func TestCacheKey_IncludesType(t *testing.T) {
	k1 := cacheKey("example.com", dns.TypeA, "resolver1")
	k2 := cacheKey("example.com", dns.TypeAAAA, "resolver1")
	require.NotEqual(t, k1, k2, "different types must not share a cache key")
}
