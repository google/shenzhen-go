// Copyright 2017 Google Inc.
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

package model

import "github.com/google/shenzhen-go/dev/source"

// Channel represents connections between pins.
type Channel struct {
	Name     string       `json:"-"`
	Type     *source.Type `json:"-"`
	Capacity int          `json:"cap"`

	// Cache of pins this channel is attached to
	Pins map[NodePin]struct{} `json:"-"`
}

// NodePin is a simple tuple of node name, pin name.
type NodePin struct{ Node, Pin string }

// AddPin is sugar for `c.Pins[NodePin{Node: node, Pin: pin}] = struct{}{}`.
// It doesn't update the node.
func (c *Channel) AddPin(node, pin string) {
	c.Pins[NodePin{Node: node, Pin: pin}] = struct{}{}
}

// HasPin is sugar for `_, found := c.Pins[NodePin{Node: node, Pin: pin}]; found`.
func (c *Channel) HasPin(node, pin string) bool {
	_, found := c.Pins[NodePin{Node: node, Pin: pin}]
	return found
}

// RemovePin is sugar for `delete(c.Pins, NodePin{Node: node, Pin: pin})`.
// It doesn't update the node.
// If the channel now has fewer than 2 pins, it can be deleted, but
// RemovePin does not do that.
func (c *Channel) RemovePin(node, pin string) {
	delete(c.Pins, NodePin{Node: node, Pin: pin})
}
