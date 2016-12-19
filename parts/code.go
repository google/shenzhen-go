// Copyright 2016 Google Inc.
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
	"shenzhen-go/graph"
	"shenzhen-go/source"
)

// Code is a component containing arbitrary code.
type Code struct {
	Code string

	// Computed from Code - which channels are read from and written to.
	chansRd, chansWr []string
}

// Channels returns the names of all channels used by this goroutine.
func (c *Code) Channels() (read, written []string) { return c.chansRd, c.chansWr }

// Impl returns the implementation of the goroutine.
func (c *Code) Impl() string { return c.Code }

// Refresh refreshes cached informatioc.
func (c *Code) Refresh(g *graph.Graph) error {
	srcs, dsts, err := source.ExtractChannelIdents(c.Code)
	if err != nil {
		return err
	}

	// Filter to only declared channels.
	c.chansWr = make([]string, 0, len(dsts))
	for _, d := range dsts {
		if _, found := g.Channels[d]; found {
			c.chansWr = append(c.chansWr, d)
		}
	}

	c.chansRd = make([]string, 0, len(srcs))
	for _, s := range srcs {
		if _, found := g.Channels[s]; found {
			c.chansRd = append(c.chansRd, s)
		}
	}
	return nil
}
