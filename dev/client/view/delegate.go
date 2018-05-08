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

import "golang.org/x/net/context"

// GraphController is implemented by the controller of a whole graph.
type GraphController interface {
	GainFocus()

	// Sub-controllers
	Nodes(func(NodeController)) // input called for all nodes
	NumNodes() int

	Channels(func(ChannelController)) // input called for all channels
	NumChannels() int

	CreateChannel(pcs ...PinController) (ChannelController, error)
	CreateNode(ctx context.Context, partType string) (NodeController, error)

	// Send properties to server
	Commit(ctx context.Context) error

	// Action links
	Save(ctx context.Context) error
	// TODO: Revert, Build, Install, Run...
}

// ChannelController is implemented by the controller of a channel.
type ChannelController interface {
	Name() string
	Pins(func(PinController)) // input called for all currently attached pins

	Attach(PinController)
	Detach(PinController)
	GainFocus()

	Commit(ctx context.Context) error
	Delete(ctx context.Context) error
}

// NodeController is implemented by the controller of a node.
type NodeController interface {
	Name() string
	Position() (x, y float64)
	Pins(func(pc PinController, channel string)) // input called for all pins on this node

	GainFocus()
	ShowMetadataSubpanel()
	ShowPartSubpanel(name string)

	Commit(ctx context.Context) error
	Delete(ctx context.Context) error
	SetPosition(ctx context.Context, x, y float64) error
}

// PinController is implemented by the controller for a pin.
type PinController interface {
	Name() string
	Type() string
	IsInput() bool
	NodeName() string
}
