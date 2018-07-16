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
	"bytes"
	"fmt"
	"text/template"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

const (
	cacheKeyTypeParam = "$Key"
	cacheCtxTypeParam = "$Ctx"
)

func cacheGetType(kt, ct string) string {
	return fmt.Sprintf("struct{ Key %s; Ctx %s }", kt, ct)
}

func cacheHitType(kt, ct string) string {
	return fmt.Sprintf("struct{ Key %s; Ctx %s; Data []byte }", kt, ct)
}

func cachePutType(kt string) string {
	return fmt.Sprintf("struct{ Key %s; Data []byte }", kt)
}

var (
	cachePins = pin.NewMap(
		&pin.Definition{
			Name:      "get",
			Direction: pin.Input,
			Type:      cacheGetType(cacheKeyTypeParam, cacheCtxTypeParam),
		},
		&pin.Definition{
			Name:      "put",
			Direction: pin.Input,
			Type:      cachePutType(cacheKeyTypeParam),
		},
		&pin.Definition{
			Name:      "hit",
			Direction: pin.Output,
			Type:      cacheHitType(cacheKeyTypeParam, cacheCtxTypeParam),
		},
		&pin.Definition{
			Name:      "miss",
			Direction: pin.Output,
			Type:      cacheGetType(cacheKeyTypeParam, cacheCtxTypeParam),
		},
	)

	cacheHeadTmpl = template.Must(template.New("cache-head").Parse(`
	const bytesLimit = {{.BytesLimit}}
	type cacheEntry struct {
		data []byte
		last time.Time
		{{if .Mult}}sync.Mutex{{end}}
	}
	{{if .Mult}}var mu sync.RWMutex{{end}}
	totalBytes := uint64(0)
	cache := make(map[{{.KeyType}}]*cacheEntry)`))

	cacheBodyTmpl = template.Must(template.New("cache-body").Parse(`
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
			{{if .Mult}}mu.RLock(){{end}}
			e, ok := cache[g.Key]
			{{if .Mult}}mu.RUnlock(){{end}}
			if !ok {
				miss <- g
				continue
			}
			{{if .Mult}}e.Lock(){{end}}
			hit <- {{.HitType}}{
				Key: g.Key,
				Ctx: g.Ctx,
				Data: e.data,
			}
			e.last = time.Now()
			{{if .Mult}}e.Unlock(){{end}}
			
		case p, open := <-put:
			if !open {
				put = nil
				continue
			}
			if len(p.Data) > bytesLimit {
				continue
			}
			
			// TODO: Can improve eviction algorithm - this is simplistic but O(n^2)
			{{if .Mult}}mu.Lock(){{end}}
			for {
				// Find something to evict if needed.
				var ek {{.KeyType}}
				var ee *cacheEntry
				et := {{.InitTime}}
				for k, e := range cache {
					{{if .Mult}}e.Lock(){{end}}
					if e.last.{{.TimeComp}}(et) {
						ee, et, ek = e, e.last, k
					}
					{{if .Mult}}e.Unlock(){{end}}
				}
				// Necessary to evict?
				if totalBytes + uint64(len(p.Data)) > bytesLimit {
					// Evict ek.
					if ee == nil {
						break
					}
					{{if .Mult}}ee.Lock(){{end}}
					totalBytes -= uint64(len(ee.data))
					{{if .Mult}}ee.Unlock(){{end}}
					delete(cache, ek)
					continue
				}

				// No - insert now.
				cache[p.Key] = &cacheEntry{
					data: p.Data,
					last: time.Now(),
				}
				totalBytes += uint64(len(p.Data))
				break
			}
			{{if .Mult}}mu.Unlock(){{end}}
		}
	}`))
)

func init() {
	model.RegisterPartType("Cache", "Flow", &model.PartType{
		New: func() model.Part {
			return &Cache{
				ContentBytesLimit: 1 << 30,
				EvictionMode:      EvictLRU,
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Cache",
				Editor: `
				<div>
					<div class="formfield">
						<label for="cache-contentbyteslimit">Maximum bytes</label>
						<input id="cache-contentbyteslimit" name="cache-contentbyteslimit" type="number" required title="Must be a whole number, at least 1." value="1073741824"></input>
					</div>
					<div class="formfield">
						<label for="cache-evictionmode">Eviction mode</label>
						<select id="cache-evictionmode" name="cache-evictionmode">
							<option value="lru" selected>LRU (least recently used)</option>
							<option value="mru">MRU (most recently used)</option>
						</select>
					</div>
				</div>`,
			},
			{
				Name: "Help",
				Editor: `<div><p>
				A Cache part caches content in memory. It supports concurrently inserting and retrieving items.
			</p><p>
				TODO: Implement Cache part.
			</p></div>`,
			},
		},
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
func (c *Cache) Impl(name string, multiple bool, types map[string]string) model.PartImpl {
	params := struct {
		KeyType, HitType, InitTime, TimeComp string
		Mult                                 bool
		BytesLimit                           uint64
	}{
		KeyType:    types[cacheKeyTypeParam],
		Mult:       multiple,
		BytesLimit: c.ContentBytesLimit,
	}
	params.HitType = cacheHitType(params.KeyType, types[cacheCtxTypeParam])
	params.InitTime, params.TimeComp = c.EvictionMode.searchParams()
	h, b := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	if err := cacheHeadTmpl.Execute(h, params); err != nil {
		panic("couldn't execute cache-head template: " + err.Error())
	}
	if err := cacheBodyTmpl.Execute(b, params); err != nil {
		panic("couldn't execute cache-body template: " + err.Error())
	}
	imps := []string{`"time"`}
	if multiple {
		imps = append(imps, `"sync"`)
	}
	return model.PartImpl{
		Imports: imps,
		Head:    h.String(),
		Body:    b.String(),
		Tail:    `close(hit); close(miss)`,
	}
}

// Pins returns a pin map.
func (c *Cache) Pins() pin.Map {
	return cachePins
}

// TypeKey returns "Cache".
func (c *Cache) TypeKey() string {
	return "Cache"
}
