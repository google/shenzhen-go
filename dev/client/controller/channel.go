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

package controller

import (
	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

type channelController struct {
	client       pb.ShenzhenGoClient
	graph        *model.Graph
	channel      *model.Channel
	existingName string
}

func (c *channelController) Name() string { return c.channel.Name }

func (c *channelController) Pins(f func(view.PinController)) {
	for p := range c.channel.Pins {
		f(&pinController{
			client: c.client,
			graph:  c.graph,
			node:   c.graph.Nodes[p.Node],
			name:   p.Pin,
		})
	}
}

func (c *channelController) Commit(ctx context.Context) error {
	np := make([]*pb.NodePin, 0, len(c.channel.Pins))
	for x := range c.channel.Pins {
		np = append(np, &pb.NodePin{
			Node: x.Node,
			Pin:  x.Pin,
		})
	}
	req := &pb.SetChannelRequest{
		Graph:   c.graph.FilePath,
		Channel: c.existingName,
		Config: &pb.ChannelConfig{
			Name: c.channel.Name,
			Type: c.channel.Type,
			Cap:  uint64(c.channel.Capacity),
			Anon: c.channel.Anonymous,
			Pins: np,
		},
	}
	_, err := c.client.SetChannel(ctx, req)
	if err != nil {
		return err // TODO: contextualise
	}
	c.existingName = c.channel.Name
	c.graph.Channels[c.channel.Name] = c.channel
	return nil
}

func (c *channelController) Delete(ctx context.Context) error {
	if c.existingName == "" {
		return nil
	}
	_, err := c.client.SetChannel(ctx, &pb.SetChannelRequest{
		Graph:   c.graph.FilePath,
		Channel: c.existingName,
	})
	return err
}

func (c *channelController) Attach(pc view.PinController) {
	n, found := c.graph.Nodes[pc.NodeName()]
	if !found {
		// TODO: return an error
		return
	}
	if _, found := n.Connections[pc.Name()]; !found {
		// TODO: return an error
	}
	n.Connections[pc.Name()] = c.Name()
	c.channel.AddPin(pc.NodeName(), pc.Name())
}

func (c *channelController) Detach(pc view.PinController) {
	n, found := c.graph.Nodes[pc.NodeName()]
	if !found {
		// TODO: return an error
		return
	}
	if _, found := n.Connections[pc.Name()]; !found {
		// TODO: return an error
	}
	n.Connections[pc.Name()] = "nil"
	c.channel.RemovePin(pc.NodeName(), pc.Name())
}
