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
	"log"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

const defaultChannelNamePrefix = "channel"

var (
	defaultChannelNameRE = regexp.MustCompile(`^channel\d+$`)

	ace = dom.GlobalAce()

	_ view.GraphController = (*graphController)(nil)
)

type graphController struct {
	doc    dom.Document
	graph  *model.Graph
	client pb.ShenzhenGoClient

	// RHS panels
	CurrentRHSPanel        dom.Element
	ChannelPropertiesPanel dom.Element
	GraphPropertiesPanel   dom.Element
	NodePropertiesPanel    dom.Element
	PreviewGoPanel         dom.Element
	PreviewJSONPanel       dom.Element
	PreviewGoSession       *dom.AceSession
	PreviewJSONSession     *dom.AceSession

	// Graph properties panel inputs
	graphNameTextInput        dom.Element
	graphPackagePathTextInput dom.Element
	graphIsCommandCheckbox    dom.Element

	// Components that are connected to whatever is selected.
	channelSharedOutlets *channelSharedOutlets
	nodeSharedOutlets    *nodeSharedOutlets
}

func setupAceView(id, mode string) *dom.AceSession {
	e := ace.Edit(id)
	if e == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	e.SetTheme(dom.AceChromeTheme)
	return e.Session().
		SetMode(mode)
}

// NewGraphController returns a new controller for a graph, and binds outlets.
func NewGraphController(doc dom.Document, graph *model.Graph, client pb.ShenzhenGoClient) view.GraphController {
	pes := make(map[string]*partEditor, len(model.PartTypes))
	for n, t := range model.PartTypes {
		p := make(map[string]dom.Element, len(t.Panels))
		for _, d := range t.Panels {
			p[d.Name] = doc.ElementByID("node-" + n + "-" + d.Name + "-panel")
		}
		pes[n] = &partEditor{
			Links:  doc.ElementByID("node-" + n + "-links"),
			Panels: p,
		}
	}

	return &graphController{
		doc:    doc,
		client: client,
		graph:  graph,

		CurrentRHSPanel: doc.ElementByID("graph-properties"),

		ChannelPropertiesPanel: doc.ElementByID("channel-properties"),
		GraphPropertiesPanel:   doc.ElementByID("graph-properties"),
		NodePropertiesPanel:    doc.ElementByID("node-properties"),
		PreviewGoPanel:         doc.ElementByID("preview-go"),
		PreviewJSONPanel:       doc.ElementByID("preview-json"),
		PreviewGoSession:       setupAceView("preview-go-ace", dom.AceGoMode),
		PreviewJSONSession:     setupAceView("preview-json-ace", dom.AceJSONMode),

		graphNameTextInput:        doc.ElementByID("graph-prop-name"),
		graphPackagePathTextInput: doc.ElementByID("graph-prop-package-path"),
		graphIsCommandCheckbox:    doc.ElementByID("graph-prop-is-command"),

		channelSharedOutlets: &channelSharedOutlets{
			inputName:     doc.ElementByID("channel-name"),
			codeType:      doc.ElementByID("channel-type"),
			inputCapacity: doc.ElementByID("channel-capacity"),
		},
		nodeSharedOutlets: &nodeSharedOutlets{
			subpanelMetadata:  doc.ElementByID("node-metadata-panel"),
			subpanelCurrent:   doc.ElementByID("node-metadata-panel"),
			inputName:         doc.ElementByID("node-name"),
			inputEnabled:      doc.ElementByID("node-enabled"),
			inputMultiplicity: doc.ElementByID("node-multiplicity"),
			inputWait:         doc.ElementByID("node-wait"),
			partEditors:       pes,
		},
	}
}

func (c *graphController) newChannelController(channel *model.Channel, existingName string) *channelController {
	return &channelController{
		client:        c.client,
		graph:         c.graph,
		channel:       channel,
		sharedOutlets: c.channelSharedOutlets,
		existingName:  existingName,
		gc:            c,
	}
}

func (c *graphController) newNodeController(node *model.Node) *nodeController {
	return &nodeController{
		client:        c.client,
		graph:         c.graph,
		node:          node,
		sharedOutlets: c.nodeSharedOutlets,
		gc:            c,
		subpanel:      c.nodeSharedOutlets.subpanelMetadata,
	}
}

func (c *graphController) GainFocus() {
	c.showRHSPanel(c.GraphPropertiesPanel)
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
		f(c.newChannelController(ch, ch.Name))
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
		Capacity: 0,
		Pins:     make(map[model.NodePin]struct{}, len(pcs)),
	}

	// Set the type to the first one found.
	// TODO: Mismatches will be reported later.
	for _, pc := range pcs {
		if ch.Type == "" {
			ch.Type = pc.Type()
		}
		ch.AddPin(pc.NodeName(), pc.Name())
	}

	// Pick a unique name
	max := -1
	for _, ec := range c.graph.Channels {
		if !defaultChannelNameRE.MatchString(ec.Name) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(ec.Name, defaultChannelNamePrefix))
		if err != nil {
			// The string just matched \d+ but can't be converted to an int?...
			panic(err)
		}
		if n > max {
			max = n
		}
	}
	ch.Name = defaultChannelNamePrefix + strconv.Itoa(max+1)

	return c.newChannelController(ch, ""), nil
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

	_, err = c.client.SetNode(ctx, &pb.SetNodeRequest{
		Graph: c.graph.FilePath,
		Config: &pb.NodeConfig{
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
	_, err := c.client.Action(ctx, &pb.ActionRequest{
		Graph:  c.graph.FilePath,
		Action: pb.ActionRequest_SAVE,
	})
	return err
}

func (c *graphController) Commit(ctx context.Context) error {
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

func (c *graphController) PreviewGo() {
	g, err := c.graph.Go()
	if err != nil {
		g = `/* Couldn't generate Go: ` + err.Error() + ` */`
	}
	c.PreviewGoSession.SetValue(g)
	c.showRHSPanel(c.PreviewGoPanel)
}

func (c *graphController) PreviewJSON() {
	g, err := c.graph.JSON()
	if err != nil {
		g = `{"error": "Couldn't generate JSON: ` + err.Error() + `"}`
	}
	c.PreviewJSONSession.SetValue(g)
	c.showRHSPanel(c.PreviewJSONPanel)
}

// showRHSPanel hides any existing panel and shows the given element as the panel.
func (c *graphController) showRHSPanel(p dom.Element) {
	if p == c.CurrentRHSPanel {
		return
	}
	c.CurrentRHSPanel.Hide()
	c.CurrentRHSPanel = p.Show()
}
