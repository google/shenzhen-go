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

package controller

import (
	"github.com/google/shenzhen-go/model"
	pb "github.com/google/shenzhen-go/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *controller) lookupGraph(graph string) (*model.Graph, error) {
	g := c.loadedGraphs[graph]
	if g == nil {
		return nil, status.Errorf(codes.NotFound, "graph %q not loaded", graph)
	}
	return g, nil
}

func (c *controller) lookupChannel(graph, channel string) (*model.Graph, *model.Channel, error) {
	g, err := c.lookupGraph(graph)
	if err != nil {
		return nil, nil, err
	}
	ch := g.Channels[channel]
	if ch == nil {
		return nil, nil, status.Errorf(codes.NotFound, "no such channel %q", channel)
	}
	return g, ch, nil
}

func (c *controller) lookupNode(graph, node string) (*model.Graph, *model.Node, error) {
	g, err := c.lookupGraph(graph)
	if err != nil {
		return nil, nil, err
	}
	n := g.Nodes[node]
	if n == nil {
		return nil, nil, status.Errorf(codes.NotFound, "no such node %q", node)
	}
	return g, n, nil
}

func (c *controller) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.Empty, error) {
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	if _, found := g.Channels[req.Name]; found {
		return &pb.Empty{}, status.Errorf(codes.FailedPrecondition, "channel %q already exists", req.Name)
	}
	// TODO: validate the name isn't silly, the type isn't silly...
	g.Channels[req.Name] = &model.Channel{
		Name:      req.Name,
		Type:      req.Type,
		Anonymous: req.Anon,
		Capacity:  int(req.Cap),
	}
	return &pb.Empty{}, nil
}

func (c *controller) ConnectPin(ctx context.Context, req *pb.ConnectPinRequest) (*pb.Empty, error) {
	_, n, err := c.lookupNode(req.Graph, req.Node)
	if err != nil {
		return &pb.Empty{}, err
	}
	_, ch, err := c.lookupChannel(req.Graph, req.Channel)
	if err != nil {
		return &pb.Empty{}, err
	}
	pin, found := n.Pins()[req.Pin]
	if !found {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	if ch.Type != pin.Type {
		return &pb.Empty{}, status.Errorf(codes.FailedPrecondition, "pin %q, channel %q type mismatch [%q != %q]", req.Pin, req.Channel, pin.Type, ch.Type)
	}
	n.Connections[req.Pin] = req.Channel
	return &pb.Empty{}, nil
}

func (c *controller) DeleteChannel(ctx context.Context, req *pb.DeleteChannelRequest) (*pb.Empty, error) {
	g, _, err := c.lookupChannel(req.Graph, req.Channel)
	if err != nil {
		return &pb.Empty{}, err
	}
	delete(g.Channels, req.Channel)
	return &pb.Empty{}, nil
}

func (c *controller) DisconnectPin(ctx context.Context, req *pb.DisconnectPinRequest) (*pb.Empty, error) {
	_, n, err := c.lookupNode(req.Graph, req.Node)
	if err != nil {
		return &pb.Empty{}, err
	}
	if _, found := n.Connections[req.Pin]; !found {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	n.Connections[req.Pin] = "nil"
	return &pb.Empty{}, nil
}

func (c *controller) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	return &pb.Empty{}, SaveJSONFile(g)
}

func (c *controller) SetGraphProperties(ctx context.Context, req *pb.SetGraphPropertiesRequest) (*pb.Empty, error) {
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Name = req.Name
	g.PackagePath = req.PackagePath
	g.IsCommand = req.IsCommand
	return &pb.Empty{}, nil
}

func (c *controller) SetNodeProperties(ctx context.Context, req *pb.SetNodePropertiesRequest) (*pb.Empty, error) {
	g, n, err := c.lookupNode(req.Graph, req.Node)
	if err != nil {
		return &pb.Empty{}, err
	}
	p, err := (&model.PartJSON{
		Part: req.PartCfg,
		Type: req.PartType,
	}).Unmarshal()
	if err != nil {
		return &pb.Empty{}, err
	}
	if n.Name != req.Name {
		if _, exists := g.Nodes[req.Name]; exists {
			return &pb.Empty{}, status.Errorf(codes.FailedPrecondition, "node %q already exists", req.Name)
		}
		delete(g.Nodes, n.Name)
		n.Name = req.Name
		g.Nodes[n.Name] = n
	}
	n.Multiplicity = uint(req.Multiplicity)
	n.Enabled = req.Enabled
	n.Wait = req.Wait
	n.Part = p
	return &pb.Empty{}, nil
}

func (c *controller) SetPosition(ctx context.Context, req *pb.SetPositionRequest) (*pb.Empty, error) {
	_, n, err := c.lookupNode(req.Graph, req.Node)
	if err != nil {
		return &pb.Empty{}, err
	}
	n.X, n.Y = int(req.X), int(req.Y)
	return &pb.Empty{}, nil
}
