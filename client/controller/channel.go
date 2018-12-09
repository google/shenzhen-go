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
	"context"

	"github.com/google/shenzhen-go/client/view"
	"github.com/google/shenzhen-go/dom"
	"github.com/google/shenzhen-go/model"
	pb "github.com/google/shenzhen-go/proto/js"
)

type channelSharedOutlets struct {
	// Channel properties inputs & outputs
	inputName     dom.Element
	codeType      dom.Element
	inputCapacity dom.Element
}

type channelController struct {
	client        pb.ShenzhenGoClient
	graph         *model.Graph
	channel       *model.Channel
	sharedOutlets *channelSharedOutlets
	existingName  string

	gc *graphController
}

func (c *channelController) Name() string { return c.channel.Name }

func (c *channelController) Pins(f func(view.PinController)) {
	for p := range c.channel.Pins {
		node := c.graph.Nodes[p.Node]
		f(&pinController{
			client: c.client,
			graph:  c.graph,
			node:   node,
			name:   p.Pin,
			def:    node.Part.Pins()[p.Pin],
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
	cfg := &pb.ChannelConfig{
		Name: c.sharedOutlets.inputName.Get("value").String(),
		Cap:  uint64(c.sharedOutlets.inputCapacity.Get("value").Int()),
		Pins: np,
	}
	req := &pb.SetChannelRequest{
		Graph:   c.graph.FilePath,
		Channel: c.existingName,
		Config:  cfg,
	}
	_, err := c.client.SetChannel(ctx, req)
	if err != nil {
		return err // TODO: contextualise
	}
	if cfg.Name != c.existingName {
		delete(c.graph.Channels, c.existingName)
		c.channel.Name = cfg.Name
		c.existingName = cfg.Name
		c.graph.Channels[cfg.Name] = c.channel
		for np := range c.channel.Pins {
			c.graph.Nodes[np.Node].Connections[np.Pin] = cfg.Name
		}
	}
	c.channel.Capacity = int(cfg.Cap)
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
	if err != nil {
		return err // TODO: contextualise
	}
	c.graph.DeleteChannel(c.channel)
	return nil
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

func (c *channelController) GainFocus() {
	c.gc.showRHSPanel(c.gc.channelPropertiesPanel)

	c.sharedOutlets.inputName.Set("value", c.channel.Name)
	c.sharedOutlets.inputCapacity.Set("value", c.channel.Capacity)
	c.sharedOutlets.codeType.Set("innerText", c.channel.Type.String())
}
