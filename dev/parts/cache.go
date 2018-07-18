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
	cache := make(map[{{.KeyType}}]*cacheEntry)
	{{if .Prometheus -}}
	cacheLimit.With(prometheus.Labels{"node_name":"{{.NodeName}}"}).Set(bytesLimit)
	cacheSize := cacheSize.With(prometheus.Labels{"node_name":"{{.NodeName}}"})
	cacheSize.Set(0)
	{{end -}}`))

	cacheBodyTmpl = template.Must(template.New("cache-body").Parse(`
	{{if .Prometheus -}}
	labels := prometheus.Labels{
		"node_name": "{{.NodeName}}",
		"instance_num": strconv.Itoa(instanceNumber),
	}
	cacheHits := cacheHits.With(labels)
	cacheMisses := cacheMisses.With(labels)
	cachePuts := cachePuts.With(labels)
	cacheEvictions := cacheEvictions.With(labels)
	cacheHitsSize := cacheHitsSize.With(labels)
	cachePutsSize := cachePutsSize.With(labels)
	cacheEvictionsSize := cacheEvictionsSize.With(labels)
	{{end -}}
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
				{{if .Prometheus}}cacheMisses.Inc(){{end}}
				continue
			}
			{{if .Mult}}e.Lock(){{end}}
			{{if .Prometheus -}}
			cacheHits.Inc()
			cacheHitsSize.Add(float64(len(e.data)))
			{{end -}}
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
					size := uint64(len(ee.data))
					{{if .Mult}}ee.Unlock(){{end}}
					totalBytes -= size
					delete(cache, ek)
					{{if .Prometheus -}}
					cacheEvictions.Inc()
					cacheEvictionsSize.Add(float64(size))
					{{end -}}
					continue
				}

				// No - insert now.
				size := uint64(len(p.Data))
				cache[p.Key] = &cacheEntry{
					data: p.Data,
					last: time.Now(),
				}
				totalBytes += size
				{{if .Prometheus -}}
				cachePuts.Inc()
				cachePutsSize.Add(float64(size))
				cacheSize.Set(float64(totalBytes))
				{{end -}}
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
		Init: `
		var (
			cacheHits = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "hits",
					Help:      "Hits to the cache in a Cache node",
				},
				[]string{"node_name", "instance_num"},
			)
			cacheMisses = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "misses",
					Help:      "Misses to the cache in a Cache node",
				},
				[]string{"node_name", "instance_num"},
			)
			cachePuts = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "puts",
					Help:      "Cache node cache insertions",
				},
				[]string{"node_name", "instance_num"},
			)
			cacheEvictions = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "evictions",
					Help:      "Cache node cache evictions",
				},
				[]string{"node_name", "instance_num"},
			)
			cacheSize = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "size",
					Help:      "Size of content in Cache nodes in bytes",
				},
				[]string{"node_name"},
			)
			cacheLimit = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "limit",
					Help:      "Upper limit of content size in Cache nodes in bytes",
				},
				[]string{"node_name"},
			)
			cacheHitsSize = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "hits_size",
					Help:      "Cumulative Cache node cache hits size in bytes",
				},
				[]string{"node_name", "instance_num"},
			)
			cachePutsSize = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "puts_size",
					Help:      "Cumulative Cache node cache insertions size in bytes",
				},
				[]string{"node_name", "instance_num"},
			)
			cacheEvictionsSize = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "shenzhen_go",
					Subsystem: "cache",
					Name:      "evictions_size",
					Help:      "Cumulative Cache node cache evictions size in bytes",
				},
				[]string{"node_name", "instance_num"},
			)
		)

		func init() {
			prometheus.MustRegister(
				cacheHits, 
				cacheMisses, 
				cachePuts, 
				cacheSize, 
				cacheLimit, 
				cacheHitsSize, 
				cachePutsSize, 
				cacheEvictionsSize,
			)
		}
		`,
		Panels: []model.PartPanel{
			{
				Name: "Cache",
				Editor: `
				<div class="form">
					<div class="formfield">
						<input id="cache-enableprometheus" name="cache-enableprometheus" type="checkbox"></input>
						<label for="cache-enableprometheus">Enable Prometheus metrics</label>
					</div>
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
	ContentBytesLimit uint64            `json:"content_bytes_limit"`
	EnablePrometheus  bool              `json:"enable_prometheus"`
	EvictionMode      CacheEvictionMode `json:"eviction_mode"`
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
func (c *Cache) Impl(n *model.Node) model.PartImpl {
	params := struct {
		BytesLimit                           uint64
		KeyType, HitType, InitTime, TimeComp string
		Mult, Prometheus                     bool
		NodeName                             string
	}{
		BytesLimit: c.ContentBytesLimit,
		KeyType:    n.TypeParams[cacheKeyTypeParam].String(),
		Mult:       n.Multiplicity != "1",
		NodeName:   n.Name,
		Prometheus: c.EnablePrometheus,
	}
	params.HitType = cacheHitType(params.KeyType, n.TypeParams[cacheCtxTypeParam].String())
	params.InitTime, params.TimeComp = c.EvictionMode.searchParams()
	h, b := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	if err := cacheHeadTmpl.Execute(h, params); err != nil {
		panic("couldn't execute cache-head template: " + err.Error())
	}
	if err := cacheBodyTmpl.Execute(b, params); err != nil {
		panic("couldn't execute cache-body template: " + err.Error())
	}
	imps := []string{`"time"`}
	if params.Mult {
		imps = append(imps, `"sync"`)
	}
	if c.EnablePrometheus {
		imps = append(imps,
			`"strconv"`,
			`"github.com/prometheus/client_golang/prometheus"`,
		)
	}
	return model.PartImpl{
		Imports:   imps,
		Head:      h.String(),
		Body:      b.String(),
		Tail:      `close(hit); close(miss)`,
		NeedsInit: c.EnablePrometheus,
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
