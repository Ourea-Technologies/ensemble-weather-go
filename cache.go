// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Cache stores responses for ETag revalidation. The client's default is an
// in-memory LRU ([NewMemoryCache] with 256 entries and 32 MiB). Pass your own
// implementation to [WithCache] to share a cache, for example across
// processes, or nil to disable caching.
//
// Keys are opaque strings built from the full request URL and a hash of the
// API key the request carried (never the key itself), so keyless and keyed
// entries never mix. Implementations must be safe for concurrent use and must
// not modify a stored entry. A failing store should behave as a miss.
type Cache interface {
	Get(ctx context.Context, key string) (*CacheEntry, bool)
	Set(ctx context.Context, key string, entry *CacheEntry)
	Delete(ctx context.Context, key string)
}

// CacheEntry is one stored response.
type CacheEntry struct {
	// ETag is the validator exactly as the server sent it (including any W/).
	ETag string
	// Body is the response body.
	Body []byte
	// Header is the response's header.
	Header http.Header
	// StoredAt is when the response was stored or last revalidated.
	StoredAt time.Time
	// MaxAge is the response's Cache-Control max-age (the API sends 0, so
	// every use revalidates).
	MaxAge time.Duration
}

func (e *CacheEntry) size() int64 {
	n := int64(len(e.Body) + len(e.ETag))
	for k, vs := range e.Header {
		n += int64(len(k))
		for _, v := range vs {
			n += int64(len(v))
		}
	}
	return n
}

// Default bounds of the in-memory cache.
const (
	DefaultCacheEntries = 256
	DefaultCacheBytes   = 32 << 20
)

// MemoryCache is a size-bounded, concurrency-safe in-memory LRU [Cache].
type MemoryCache struct {
	mu         sync.Mutex
	maxEntries int
	maxBytes   int64
	bytes      int64
	order      *list.List // front = most recently used
	items      map[string]*list.Element
}

type memoryItem struct {
	key   string
	entry *CacheEntry
	size  int64
}

// NewMemoryCache returns an LRU cache bounded to maxEntries entries and
// maxBytes bytes. A bound of 0 or less selects the default.
func NewMemoryCache(maxEntries int, maxBytes int64) *MemoryCache {
	if maxEntries <= 0 {
		maxEntries = DefaultCacheEntries
	}
	if maxBytes <= 0 {
		maxBytes = DefaultCacheBytes
	}
	return &MemoryCache{maxEntries: maxEntries, maxBytes: maxBytes, order: list.New(), items: map[string]*list.Element{}}
}

// Get returns the entry stored under key.
func (c *MemoryCache) Get(_ context.Context, key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*memoryItem).entry, true
}

// Set stores entry under key, evicting the least recently used entries to
// stay within the bounds. An entry larger than the byte bound is not stored.
func (c *MemoryCache) Set(_ context.Context, key string, entry *CacheEntry) {
	size := entry.size() + int64(len(key))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(key)
	if size > c.maxBytes {
		return
	}
	c.items[key] = c.order.PushFront(&memoryItem{key: key, entry: entry, size: size})
	c.bytes += size
	for c.order.Len() > c.maxEntries || c.bytes > c.maxBytes {
		c.remove(c.order.Back().Value.(*memoryItem).key)
	}
}

// Delete removes the entry stored under key.
func (c *MemoryCache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(key)
}

// Len returns the number of stored entries.
func (c *MemoryCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

func (c *MemoryCache) remove(key string) {
	if el, ok := c.items[key]; ok {
		c.bytes -= el.Value.(*memoryItem).size
		c.order.Remove(el)
		delete(c.items, key)
	}
}

// cacheKey is the request URL plus a hash of the credential it carried.
func cacheKey(url, apiKey string) string {
	if apiKey == "" {
		return url + " anonymous"
	}
	sum := sha256.Sum256([]byte(apiKey))
	return url + " key:" + hex.EncodeToString(sum[:8])
}

type cacheControl struct {
	noStore bool
	noCache bool
	maxAge  time.Duration
}

func parseCacheControl(h string) cacheControl {
	var cc cacheControl
	for _, part := range strings.Split(h, ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch strings.ToLower(name) {
		case "no-store":
			cc.noStore = true
		case "no-cache":
			cc.noCache = true
		case "max-age":
			if n, err := strconv.Atoi(strings.Trim(value, `"`)); err == nil && n > 0 {
				cc.maxAge = time.Duration(n) * time.Second
			}
		}
	}
	if cc.noCache {
		cc.maxAge = 0
	}
	return cc
}
