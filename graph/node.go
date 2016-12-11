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

import "shenzhen-go/source"

// Node models a goroutine.
type Node struct {
	Name, Code string
	Wait       bool

	// Computed from Code - which channels are read from and written to.
	chansRd, chansWr []string
}

// ChannelsRead returns the names of all channels read by this goroutine.
func (n *Node) ChannelsRead() []string { return n.chansRd }

// ChannelsWritten returns the names of all channels written by this goroutine.
func (n *Node) ChannelsWritten() []string { return n.chansWr }

// UpdateChans refreshes the channels known to be used by the goroutine.
func (n *Node) UpdateChans(chans map[string]*Channel) error {
	srcs, dsts, err := source.ExtractChannelIdents(n.Code)
	if err != nil {
		return err
	}
	// dsts is definitely all the channels written by this goroutine.
	n.chansWr = dsts

	// srcs can include false positives, so filter them.
	n.chansRd = make([]string, 0, len(srcs))
	for _, s := range srcs {
		if _, found := chans[s]; found {
			n.chansRd = append(n.chansRd, s)
		}
	}
	return nil
}

func (n *Node) String() string { return n.Name }
