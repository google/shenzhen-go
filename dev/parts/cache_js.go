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

//+build js

package parts

import "github.com/google/shenzhen-go/dev/dom"

var (
	inputCacheContentBytesLimit = doc.ElementByID("cache-contentbyteslimit")
	selectCacheEvictionMode     = doc.ElementByID("cache-evictionmode")

	focusedCache *Cache
)

func init() {
	inputCacheContentBytesLimit.AddEventListener("change", func(dom.Object) {
		focusedCache.ContentBytesLimit = inputCacheContentBytesLimit.Get("value").Uint64()
	})
	selectCacheEvictionMode.AddEventListener("change", func(dom.Object) {
		focusedCache.EvictionMode = CacheEvictionMode(selectCacheEvictionMode.Get("value").String())
	})
}

func (c *Cache) GainFocus() {
	focusedCache = c
	inputCacheContentBytesLimit.Set("value", c.ContentBytesLimit)
	selectCacheEvictionMode.Set("value", c.EvictionMode)
}
