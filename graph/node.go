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

package graph

import (
	"encoding/json"
	"fmt"

	"shenzhen-go/parts"
)

// While being developed, check the interface is matched.
var _ = Part(&parts.Multiplexer{})

// Part abstracts the implementation of a node.
type Part interface {
	Channels() (read, written []string)
	Impl() string
	Refresh() error
	TypeKey() string
}

// Node models a goroutine.
type Node struct {
	Part

	Name string
	Wait bool
}

// ChannelsRead returns the channels read from by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsRead() []string {
	r, _ := n.Part.Channels()
	return r
}

// ChannelsWritten returns the channels written to by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsWritten() []string {
	_, w := n.Part.Channels()
	return w
}

func (n *Node) String() string { return n.Name }

type jsonNode struct {
	Name     string          `json:"name"`
	Wait     bool            `json:"wait"`
	Part     json.RawMessage `json:"part"`
	PartType string          `json:"part_type"`
}

// MarshalJSON encodes the node and part as JSON.
func (n *Node) MarshalJSON() ([]byte, error) {
	p, err := json.Marshal(n.Part)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&jsonNode{
		Part:     p,
		PartType: n.Part.TypeKey(),
		Name:     n.Name,
		Wait:     n.Wait,
	})
}

// UnmarshalJSON decodes the node and part as JSON.
func (n *Node) UnmarshalJSON(j []byte) error {
	var mp jsonNode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	pf, ok := parts.PartFactories[mp.PartType]
	if !ok {
		return fmt.Errorf("unknown part type %q", mp.PartType)
	}
	p := pf()
	if err := json.Unmarshal(mp.Part, p); err != nil {
		return err
	}
	ip, ok := p.(Part)
	if !ok {
		return fmt.Errorf("unmarshalled to a non-part [%T !~ Part]", p)
	}
	n.Name = mp.Name
	n.Wait = mp.Wait
	n.Part = ip
	return nil
}
