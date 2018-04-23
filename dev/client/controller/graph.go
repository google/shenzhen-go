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

	// RHS panels
	CurrentRHSPanel      dom.Element
	GraphPropertiesPanel dom.Element
	NodePropertiesPanel  dom.Element
	// ChannelPropertiesPanel dom.Element

	// Graph properties panel inputs
	graphNameTextInput        dom.Element
	graphPackagePathTextInput dom.Element
	graphIsCommandCheckbox    dom.Element

	// Shared outlets for nodes, stored here for passing to new nodeControllers
	nodeSharedOutlets *nodeSharedOutlets
}

// NewGraphController returns a new controller for a graph, and binds outlets.
func NewGraphController(doc dom.Document, graph *model.Graph, client pb.ShenzhenGoClient) view.GraphController {
	nso := &nodeSharedOutlets{
		nodeMetadataSubpanel:  doc.ElementByID("node-metadata-panel"),
		nodeCurrentSubpanel:   doc.ElementByID("node-metadata-panel"),
		nodeNameInput:         doc.ElementByID("node-name"),
		nodeEnabledInput:      doc.ElementByID("node-enabled"),
		nodeMultiplicityInput: doc.ElementByID("node-multiplicity"),
		nodeWaitInput:         doc.ElementByID("node-wait"),
		nodePartEditors:       make(map[string]*partEditor, len(model.PartTypes)),
	}

	for n, t := range model.PartTypes {
		p := make(map[string]dom.Element, len(t.Panels))
		for _, d := range t.Panels {
			p[d.Name] = doc.ElementByID("node-" + n + "-" + d.Name + "-panel")
		}
		nso.nodePartEditors[n] = &partEditor{
			Links:  doc.ElementByID("node-" + n + "-links"),
			Panels: p,
		}
	}

	return &graphController{
		doc:    doc,
		client: client,
		graph:  graph,

		GraphPropertiesPanel: doc.ElementByID("graph-properties"),
		NodePropertiesPanel:  doc.ElementByID("node-properties"),
		CurrentRHSPanel:      doc.ElementByID("graph-properties"),

		graphNameTextInput:        doc.ElementByID("graph-prop-name"),
		graphPackagePathTextInput: doc.ElementByID("graph-prop-package-path"),
		graphIsCommandCheckbox:    doc.ElementByID("graph-prop-is-command"),

		nodeSharedOutlets: nso,
	}
}

func (c *graphController) newChannelController(channel *model.Channel) *channelController {
	return &channelController{
		client:  c.client,
		graph:   c.graph,
		channel: channel,
	}
}

func (c *graphController) newNodeController(node *model.Node) *nodeController {
	return &nodeController{
		client:        c.client,
		graph:         c.graph,
		node:          node,
		sharedOutlets: c.nodeSharedOutlets,
		gc:            c,
		subpanel:      c.nodeSharedOutlets.nodeMetadataSubpanel,
	}
}

func (c *graphController) GainFocus() {
	c.showRHSPanel(c.GraphPropertiesPanel)
}

func (c *graphController) LoseFocus() {
	// Nop.
}

func (c *graphController) Nodes(f func(view.NodeController)) {
	for _, n := range c.graph.Nodes {
		f(c.newNodeController(n))
	}
}

func (c *graphController) NumNodes() int {
	return len(c.graph.Nodes)
}

func (c *graphController) Channels(f func(view.ChannelController)) {
	for _, ch := range c.graph.Channels {
		f(c.newChannelController(ch))
	}
}

func (c *graphController) NumChannels() int {
	return len(c.graph.Channels)
}

func (c graphController) PartTypes() map[string]*model.PartType {
	return model.PartTypes
}

func (c *graphController) CreateChannel(pcs ...view.PinController) (view.ChannelController, error) {
	ch := &model.Channel{
		Capacity:  0,
		Anonymous: true,
		Pins:      make(map[model.NodePin]struct{}, len(pcs)),
	}

	// Set the type to the first one found.
	// TODO: Mismatches will be reported later.
	for _, pc := range pcs {
		if ch.Type == "" {
			ch.Type = pc.Type()
		}
		ch.Pins[model.NodePin{Node: pc.NodeName(), Pin: pc.Name()}] = struct{}{}
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
	return c.newNodeController(n), nil
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

// showRHSPanel hides any existing panel and shows the given element as the panel.
func (c *graphController) showRHSPanel(p dom.Element) {
	if p == c.CurrentRHSPanel {
		return
	}
	c.CurrentRHSPanel.Hide()
	c.CurrentRHSPanel = p.Show()
}
