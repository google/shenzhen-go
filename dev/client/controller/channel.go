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
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
	"golang.org/x/net/context"
)

type channelController struct {
	client  pb.ShenzhenGoClient
	graph   *model.Graph
	channel *model.Channel
	created bool
}

func (c *channelController) Channel() *model.Channel {
	return c.channel
}

func (c *channelController) Attach(node, pin string) error {
	// TODO: implement
	return nil
}

func (c *channelController) Detach(node, pin string) error {
	// TODO: implement
	return nil
}

func (c *channelController) Commit(ctx context.Context) error {
	if c.created {
		// TODO: implement update
		return nil
	}

	if err := c.create(ctx); err != nil {
		return err
	}
	c.graph.Channels[c.channel.Name] = c.channel
	c.created = true
	return nil
}

func (c *channelController) Delete(ctx context.Context) error {
	// TODO: implement
	return nil
}

func (c *channelController) create(ctx context.Context) error {
	np := make([]*pb.NodePin, 0, len(c.channel.Pins))
	for x := range c.channel.Pins {
		np = append(np, &pb.NodePin{
			Node: x.Node,
			Pin:  x.Pin,
		})
	}
	req := &pb.CreateChannelRequest{
		Graph: c.graph.FilePath,
		Name:  c.channel.Name,
		Type:  c.channel.Type,
		Cap:   uint64(c.channel.Capacity),
		Anon:  c.channel.Anonymous,
		Pins:  np,
	}
	_, err := c.client.CreateChannel(ctx, req)
	return err
}
