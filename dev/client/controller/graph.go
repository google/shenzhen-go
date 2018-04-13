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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

const anonChannelNamePrefix = "anonymousChannel"

var anonChannelNameRE = regexp.MustCompile(`^anonymousChannel\d+$`)

type graphController struct {
	doc    dom.Document
	graph  *model.Graph
	client pb.ShenzhenGoClient

	// Graph properties panel inputs
	graphNameTextInput        dom.Element
	graphPackagePathTextInput dom.Element
	graphIsCommandCheckbox    dom.Element
}

// NewGraphController returns a new controller for a graph and binds outlets.
func NewGraphController(d dom.Document, g *model.Graph, c pb.ShenzhenGoClient) view.GraphController {
	return &graphController{
		doc:    d,
		client: c,
		graph:  g,

		graphNameTextInput:        d.ElementByID("graph-prop-name"),
		graphPackagePathTextInput: d.ElementByID("graph-prop-package-path"),
		graphIsCommandCheckbox:    d.ElementByID("graph-prop-is-command"),
	}
}

func (c *graphController) Graph() *model.Graph {
	return c.graph
}

func (c *graphController) Nodes(f func(view.NodeController)) {
	for _, n := range c.graph.Nodes {
		f(&nodeController{node: n})
	}
}

func (c *graphController) Node(name string) view.NodeController {
	n := c.graph.Nodes[name]
	if n == nil {
		return nil
	}
	return &nodeController{node: n}
}

func (c *graphController) NumNodes() int {
	return len(c.graph.Nodes)
}

func (c *graphController) Channel(name string) view.ChannelController {
	ch := c.graph.Channels[name]
	if ch == nil {
		return nil
	}
	return &channelController{
		client:  c.client,
		graph:   c.graph,
		channel: ch,
	}
}

func (c *graphController) Channels(f func(view.ChannelController)) {
	for _, ch := range c.graph.Channels {
		f(&channelController{
			client:  c.client,
			graph:   c.graph,
			channel: ch,
		})
	}
}

func (c *graphController) NumChannels() int {
	return len(c.graph.Channels)
}

func (c graphController) PartTypes() map[string]*model.PartType {
	return model.PartTypes
}

func (c *graphController) CreateChannel(firstNode, firstPin string) (view.ChannelController, error) {
	node := c.graph.Nodes[firstNode]
	if node == nil {
		return nil, fmt.Errorf("node %q does not exist", firstNode)
	}

	pin := node.Pins()[firstPin]
	if pin == nil {
		return nil, fmt.Errorf("pin %q does not exist on node %q", firstPin, firstNode)
	}

	ch := &model.Channel{
		Type:      pin.Type,
		Capacity:  0,
		Anonymous: true,
	}
	// Pick a unique name
	max := -1
	for _, ec := range c.graph.Channels {
		if !anonChannelNameRE.MatchString(ec.Name) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(ec.Name, anonChannelNamePrefix))
		if err != nil {
			// The string just matched \d+ but can't be converted to an int?...
			panic(err)
		}
		if n > max {
			max = n
		}
	}
	ch.Name = anonChannelNamePrefix + strconv.Itoa(max+1)

	return &channelController{
		client:  c.client,
		graph:   c.graph,
		channel: ch,
	}, nil
}

func (c *graphController) CreateNode(ctx context.Context, partType string) (view.NodeController, error) {
	// Invent a reasonable unique name.
	name := partType

	for i := 2; ; i++ {
		if _, found := c.graph.Nodes[name]; !found {
			break
		}
		name = partType + " " + strconv.Itoa(i)
	}
	pt := model.PartTypes[partType].New()
	pm, err := model.MarshalPart(pt)
	if err != nil {
		return nil, errors.New("marshalling part: " + err.Error())
	}

	n := &model.Node{
		Name:         name,
		Enabled:      true,
		Wait:         true,
		Multiplicity: 1,
		Part:         pt,
		// TODO: use a better initial position
		X: 150,
		Y: 150,
	}

	_, err = c.client.CreateNode(ctx, &pb.CreateNodeRequest{
		Graph: c.graph.FilePath,
		Props: &pb.NodeConfig{
			Name:         n.Name,
			Enabled:      n.Enabled,
			Wait:         n.Wait,
			Multiplicity: uint32(n.Multiplicity),
			PartType:     partType,
			PartCfg:      pm.Part,
		},
	})
	if err != nil {
		return nil, err
	}
	c.graph.Nodes[n.Name] = n
	return &nodeController{
		client: c.client,
		graph:  c.graph,
		node:   n,
	}, nil
}

func (c *graphController) Save(ctx context.Context) error {
	_, err := c.client.Save(ctx, &pb.SaveRequest{Graph: c.graph.FilePath})
	return err
}

func (c *graphController) SaveProperties(ctx context.Context) error {
	req := &pb.SetGraphPropertiesRequest{
		Graph:       c.graph.FilePath,
		Name:        c.graphNameTextInput.Get("value").String(),
		PackagePath: c.graphPackagePathTextInput.Get("value").String(),
		IsCommand:   c.graphIsCommandCheckbox.Get("checked").Bool(),
	}
	if _, err := c.client.SetGraphProperties(ctx, req); err != nil {
		return err
	}
	c.graph.Name = req.Name
	c.graph.PackagePath = req.PackagePath
	c.graph.IsCommand = req.IsCommand
	return nil
}
