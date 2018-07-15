// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parts

import (
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

const cacheKeyTypeParam = "$Key"

func cachePutType(kt string) string {
	return fmt.Sprintf("struct{ Key %s; Data []byte }", kt)
}

var (
	cachePins = pin.NewMap(
		&pin.Definition{
			Name:      "get",
			Direction: pin.Input,
			Type:      cacheKeyTypeParam,
		},
		&pin.Definition{
			Name:      "put",
			Direction: pin.Input,
			Type:      cachePutType(cacheKeyTypeParam),
		},
		&pin.Definition{
			Name:      "hit",
			Direction: pin.Output,
			Type:      cachePutType(cacheKeyTypeParam),
		},
		&pin.Definition{
			Name:      "miss",
			Direction: pin.Output,
			Type:      cacheKeyTypeParam,
		},
	)
)

func init() {
	model.RegisterPartType("Cache", "Flow", &model.PartType{
		New: func() model.Part {
			return &Cache{
				ContentBytesLimit: 1 << 30,
				EvictionMode:      EvictLRU,
			}
		},
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div><p>
				A Cache part caches content in memory.
			</p><p>
				TODO: Implement Cache part.
			</p></div>`,
		}},
	})
}

// Cache is a part which caches content in memory.
type Cache struct {
	ContentBytesLimit uint64
	EvictionMode      CacheEvictionMode
}

// CacheEvictionMode is how the cache decides which content to evict
// to stay under the memory limit.
type CacheEvictionMode string

// Cache eviction modes.
const (
	EvictLRU CacheEvictionMode = "lru" // Least recently used
	EvictMRU CacheEvictionMode = "mru" // Most recently used
)

func (m CacheEvictionMode) searchParams() (init, comp string) {
	switch m {
	case EvictLRU:
		return "time.Now()", "Before"
	case EvictMRU:
		return "time.Time{}", "After"
	default:
		panic("unrecognised EvictionMode " + m)
	}
}

// Clone returns a clone of this Cache.
func (c *Cache) Clone() model.Part {
	c0 := *c
	return &c0
}

// Impl returns a cache implementation.
func (c *Cache) Impl(types map[string]string) (head, body, tail string) {
	keyType := types[cacheKeyTypeParam]
	putType := cachePutType(keyType)
	initTime, timeComp := c.EvictionMode.searchParams()
	return fmt.Sprintf(`
		const bytesLimit = %d
		type cacheEntry struct {
			data []byte
			last time.Time
			mu   sync.Mutex
		}
		var mu sync.RWMutex
		totalBytes := uint64(0)
		cache := make(map[%s]*cacheEntry)
	`, c.ContentBytesLimit, keyType),
		fmt.Sprintf(`
		for {
			if get == nil && put == nil {
				break
			}
			select {
			case g, open := <-get:
				if !open {
					get = nil
					continue
				}
				mu.RLock()
				e, ok := cache[g]
				mu.RUnlock()
				if !ok {
					miss <- g
					continue
				}
				e.mu.Lock()
				hit <- %s{
					Key: g,
					Data: e.data,
				}
				e.last = time.Now()
				e.mu.Unlock()
				
			case p, open := <-put:
				if !open {
					put = nil
					continue
				}
				if len(p.Data) > bytesLimit {
					// TODO: some kind of failure message
					continue
				}
				
				// TODO: Can improve eviction algorithm - this is simplistic but O(n^2)
				for {
					mu.RLock()
					if totalBytes + uint64(len(p.Data)) <= bytesLimit {
						mu.RUnlock()
						break
					}
					et := %s
					var ek %s
					for k, e := range cache {
						e.Lock()
						if e.last.%s(et) {
							et, ek = e.last, k
						}
						e.Unlock()
					}
					mu.RUnlock()
					mu.Lock()
					// Still necessary to evict?
					if newBytes := totalBytes + uint64(len(p.Data)); newBytes <= bytesLimit {
						cache[p.Key] = &cacheEntry{
							data: p.Data,
							last: time.Now(),
						}
						totalBytes = newBytes
						mu.Unlock()
						break
					}
					// Evict k.
					totalBytes -= len(cache[k].data)
					delete(cache, k)
					if newBytes := totalBytes + uint64(len(p.Data)); newBytes <= bytesLimit {
						cache[p.Key] = &cacheEntry{
							data: p.Data,
							last: time.Now(),
						}
						totalBytes = newBytes
						mu.Unlock()
						break
					}
					mu.Unlock()
				}
			}
		}`, putType, initTime, keyType, timeComp),
		`close(hit)
		close(miss)`
}

// Imports returns nil.
func (c *Cache) Imports() []string {
	return []string{`"sync"`, `"time"`}
}

// Pins returns a pin map.
func (c *Cache) Pins() pin.Map {
	return cachePins
}

// TypeKey returns "Cache".
func (c *Cache) TypeKey() string {
	return "Cache"
}
