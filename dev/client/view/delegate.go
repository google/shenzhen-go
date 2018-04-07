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

package view

import (
	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/model"
)

// GraphController is implemented by the view controller of a whole graph.
type GraphController interface {
	Graph() *model.Graph // TODO: remove
	PartTypes() map[string]*model.PartType

	// Sub-controllers
	Node(name string) NodeController
	Nodes(func(NodeController)) // input called for all nodes
	NumNodes() int
	Channel(name string) ChannelController
	Channels(func(ChannelController)) // input called for all channels
	NumChannels() int

	CreateNode(ctx context.Context, partType string) (*model.Node, error)
	Save(ctx context.Context) error
	SaveProperties(ctx context.Context) error
}

// ChannelController is implemented by the controller of a channel.
type ChannelController interface {
	Channel() *model.Channel // TODO: remove
}

// NodeController is implemented by the controller of a node.
type NodeController interface {
	Node() *model.Node // TODO: remove

	Delete() error
}
