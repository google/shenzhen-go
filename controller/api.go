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
	"fmt"
	"net/http"

	"github.com/google/shenzhen-go/api"
	"github.com/google/shenzhen-go/model"
	"golang.org/x/net/context"
)

type apiHandler struct{}

// API handles all API requests.
var API apiHandler

func lookupGraph(graph string) (*model.Graph, error) {
	g := loadedGraphs[graph]
	if g == nil {
		return nil, api.Statusf(http.StatusNotFound, "graph not loaded: %q", graph)
	}
	return g, nil
}

func lookupChannel(graph, channel string) (*model.Graph, *model.Channel, error) {
	g, err := lookupGraph(graph)
	if err != nil {
		return nil, nil, err
	}
	c := g.Channels[channel]
	if c == nil {
		return nil, nil, api.Statusf(http.StatusNotFound, "no such channel: %q", channel)
	}
	return g, c, nil
}

func lookupNode(graph, node string) (*model.Graph, *model.Node, error) {
	g, err := lookupGraph(graph)
	if err != nil {
		return nil, nil, err
	}
	n := g.Nodes[node]
	if n == nil {
		return nil, nil, api.Statusf(http.StatusNotFound, "no such node: %q", node)
	}
	return g, n, nil
}

func (h apiHandler) CreateChannel(ctx context.Context, req *api.CreateChannelRequest) (*api.Empty, error) {
	g, err := lookupGraph(req.Graph)
	if err != nil {
		return &api.Empty{}, err
	}
	if _, found := g.Channels[req.Name]; found {
		return &api.Empty{}, api.Statusf(http.StatusBadRequest, "channel %q already exists", req.Name)
	}
	// TODO: validate the name isn't silly, the type isn't silly...
	g.Channels[req.Name] = &model.Channel{
		Name:      req.Name,
		Type:      req.Type,
		Anonymous: req.Anon,
		Capacity:  int(req.Cap),
	}
	return &api.Empty{}, nil
}

func (h apiHandler) ConnectPin(ctx context.Context, req *api.ConnectPinRequest) (*api.Empty, error) {
	_, n, err := lookupNode(req.Graph, req.Node)
	if err != nil {
		return &api.Empty{}, err
	}
	if _, found := n.Connections[req.Pin]; !found {
		return &api.Empty{}, api.Statusf(http.StatusNotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	n.Connections[req.Pin] = req.Channel
	return &api.Empty{}, nil
}

func (h apiHandler) DeleteChannel(ctx context.Context, req *api.DeleteChannelRequest) (*api.Empty, error) {
	g, _, err := lookupChannel(req.Graph, req.Channel)
	if err != nil {
		return &api.Empty{}, err
	}
	delete(g.Channels, req.Channel)
	return &api.Empty{}, nil
}

func (h apiHandler) DisconnectPin(ctx context.Context, req *api.DisconnectPinRequest) (*api.Empty, error) {
	_, n, err := lookupNode(req.Graph, req.Node)
	if err != nil {
		return &api.Empty{}, err
	}
	if _, found := n.Connections[req.Pin]; !found {
		return &api.Empty{}, api.Statusf(http.StatusNotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	n.Connections[req.Pin] = "nil"
	return &api.Empty{}, nil
}

func (h apiHandler) Save(ctx context.Context, req *api.SaveRequest) (*api.Empty, error) {
	g, err := lookupGraph(req.Graph)
	if err != nil {
		return &api.Empty{}, err
	}
	return &api.Empty{}, SaveJSONFile(g)
}

func (h apiHandler) SetGraphProperties(ctx context.Context, req *api.SetGraphPropertiesRequest) (*api.Empty, error) {
	g, err := lookupGraph(req.Graph)
	if err != nil {
		return &api.Empty{}, err
	}
	g.Name = req.Name
	g.PackagePath = req.PackagePath
	g.IsCommand = req.IsCommand
	return &api.Empty{}, nil
}

func (h apiHandler) SetNodeProperties(ctx context.Context, req *api.SetNodePropertiesRequest) (*api.Empty, error) {
	g, n, err := lookupNode(req.Graph, req.Node)
	if err != nil {
		return &api.Empty{}, err
	}
	p, err := (&model.PartJSON{
		Part: req.PartCfg,
		Type: req.PartType,
	}).Unmarshal()
	if err != nil {
		return &api.Empty{}, err
	}
	if n.Name != req.Name {
		if _, exists := g.Nodes[req.Name]; exists {
			return &api.Empty{}, fmt.Errorf("node %q already exists", req.Name)
		}
		delete(g.Nodes, n.Name)
		n.Name = req.Name
		g.Nodes[n.Name] = n
	}
	n.Multiplicity = uint(req.Multiplicity)
	n.Enabled = req.Enabled
	n.Wait = req.Wait
	n.Part = p
	return &api.Empty{}, nil
}

func (h apiHandler) SetPosition(ctx context.Context, req *api.SetPositionRequest) (*api.Empty, error) {
	_, n, err := lookupNode(req.Graph, req.Node)
	if err != nil {
		return &api.Empty{}, err
	}
	n.X, n.Y = int(req.X), int(req.Y)
	return &api.Empty{}, nil
}
