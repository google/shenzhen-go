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

// Package server serves the user interface and API, and manages the data model.
package server

import (
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/dev/model"
)

// S is the server singleton.
var S = &server{
	loadedGraphs: make(map[string]*serveGraph),
}

type server struct {
	loadedGraphs map[string]*serveGraph
	sync.Mutex
}

func (c *server) lookupGraph(key string) (*serveGraph, error) {
	c.Lock()
	defer c.Unlock()
	g := c.loadedGraphs[key]
	if g == nil {
		return nil, status.Errorf(codes.NotFound, "graph %q not loaded", key)
	}
	return g, nil
}

func (c *server) createGraph(key string, graph *model.Graph) (*serveGraph, error) {
	c.Lock()
	defer c.Unlock()
	if c.loadedGraphs[key] != nil {
		return nil, status.Errorf(codes.NotFound, "graph %q already created", key)
	}
	sg := &serveGraph{Graph: graph}
	c.loadedGraphs[key] = sg
	return sg, nil
}

type serveGraph struct {
	*model.Graph
	sync.Mutex
}

func (g *serveGraph) lookupChannel(channel string) (*model.Channel, error) {
	ch := g.Channels[channel]
	if ch == nil {
		return nil, status.Errorf(codes.NotFound, "no such channel %q", channel)
	}
	return ch, nil
}

func (g *serveGraph) lookupNode(node string) (*model.Node, error) {
	n := g.Nodes[node]
	if n == nil {
		return nil, status.Errorf(codes.NotFound, "no such node %q", node)
	}
	return n, nil
}
