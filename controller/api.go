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
	"log"
	"net/http"

	"github.com/google/shenzhen-go/api"
	"github.com/google/shenzhen-go/model"
)

type apiHandler struct{}

// API handles all API requests.
var API apiHandler

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s api: %s", r.Method, r.URL)
	api.Dispatch(h, w, r)
}

func lookupGraph(req *api.Request) (*model.Graph, error) {
	g := loadedGraphs[req.Graph]
	if g == nil {
		return nil, api.Statusf(http.StatusNotFound, "graph not loaded: %q", req.Graph)
	}
	return g, nil
}

func lookupChannel(req *api.ChannelRequest) (*model.Graph, *model.Channel, error) {
	g, err := lookupGraph(&req.Request)
	if err != nil {
		return nil, nil, err
	}
	c := g.Channels[req.Channel]
	if c == nil {
		return nil, nil, api.Statusf(http.StatusNotFound, "no such channel: %q", req.Channel)
	}
	return g, c, nil
}

func lookupNode(req *api.NodeRequest) (*model.Graph, *model.Node, error) {
	g, err := lookupGraph(&req.Request)
	if err != nil {
		return nil, nil, err
	}
	n := g.Nodes[req.Node]
	if n == nil {
		return nil, nil, api.Statusf(http.StatusNotFound, "no such node: %q", req.Node)
	}
	return g, n, nil
}

func (h apiHandler) CreateChannel(req *api.CreateChannelRequest) error {
	g, err := lookupGraph(&req.Request)
	if err != nil {
		return err
	}
	if _, found := g.Channels[req.Name]; found {
		return api.Statusf(http.StatusBadRequest, "channel %q already exists", req.Name)
	}
	// TODO: validate the name isn't silly, the type isn't silly...
	g.Channels[req.Name] = &model.Channel{
		Name:      req.Name,
		Type:      req.Type,
		Anonymous: req.Anonymous,
		Capacity:  req.Capacity,
	}
	return nil
}

func (h apiHandler) ConnectPin(req *api.ConnectPinRequest) error {
	_, n, err := lookupNode(&req.NodeRequest)
	if err != nil {
		return err
	}
	if _, found := n.Connections[req.Pin]; !found {
		return api.Statusf(http.StatusNotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	n.Connections[req.Pin] = req.Channel
	return nil
}

func (h apiHandler) DeleteChannel(req *api.ChannelRequest) error {
	g, _, err := lookupChannel(req)
	if err != nil {
		return err
	}
	delete(g.Channels, req.Channel)
	return nil
}

func (h apiHandler) DisconnectPin(req *api.PinRequest) error {
	_, n, err := lookupNode(&req.NodeRequest)
	if err != nil {
		return err
	}
	if _, found := n.Connections[req.Pin]; !found {
		return api.Statusf(http.StatusNotFound, "no pin %q on node %q", req.Pin, req.Node)
	}
	n.Connections[req.Pin] = "nil"
	return nil
}

func (h apiHandler) Save(req *api.Request) error {
	g, err := lookupGraph(req)
	if err != nil {
		return err
	}
	return SaveJSONFile(g)
}

func (h apiHandler) SetGraphProperties(req *api.SetGraphPropertiesRequest) error {
	g, err := lookupGraph(&req.Request)
	if err != nil {
		return err
	}
	g.Name = req.Name
	g.PackagePath = req.PackagePath
	g.IsCommand = req.IsCommand
	return nil
}

func (h apiHandler) SetNodeProperties(req *api.SetNodePropertiesRequest) error {
	g, n, err := lookupNode(&req.NodeRequest)
	if err != nil {
		return err
	}
	if n.Name != req.Name {
		if _, exists := g.Nodes[req.Name]; exists {
			return fmt.Errorf("node %q already exists", req.Name)
		}
		delete(g.Nodes, n.Name)
		n.Name = req.Name
		g.Nodes[n.Name] = n
	}
	n.Multiplicity = req.Multiplicity
	n.Enabled = req.Enabled
	n.Wait = req.Wait

	return nil
}

func (h apiHandler) SetPosition(req *api.SetPositionRequest) error {
	_, n, err := lookupNode(&req.NodeRequest)
	if err != nil {
		return err
	}
	n.X, n.Y = req.X, req.Y
	return nil
}
