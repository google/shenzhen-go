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
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

type partEditor struct {
	Links  dom.Element
	Panels map[string]dom.Element
}

type nodeSharedOutlets struct {
	// Node properties subpanels and inputs
	nodeMetadataSubpanel  dom.Element
	nodeCurrentSubpanel   dom.Element
	nodeNameInput         dom.Element
	nodeEnabledInput      dom.Element
	nodeMultiplicityInput dom.Element
	nodeWaitInput         dom.Element
	nodePartEditors       map[string]*partEditor
}

type nodeController struct {
	client        pb.ShenzhenGoClient
	graph         *model.Graph
	node          *model.Node
	sharedOutlets *nodeSharedOutlets

	gc       *graphController
	subpanel dom.Element // remember most recent subpanel for each node
}

func (c *nodeController) Name() string             { return c.node.Name }
func (c *nodeController) Position() (x, y float64) { return c.node.X, c.node.Y }

func (c *nodeController) Pins(f func(view.PinController)) {
	for name := range c.node.Pins() {
		f(&pinController{
			client: c.client,
			graph:  c.graph,
			node:   c.node,
			name:   name,
		})
	}
}

func (c *nodeController) Delete(ctx context.Context) error {
	_, err := c.client.DeleteNode(ctx, &pb.DeleteNodeRequest{
		Graph: c.graph.FilePath,
		Node:  c.node.Name,
	})
	return err // TODO: contextualise
}

func (c *nodeController) Save(ctx context.Context) error {
	pj, err := model.MarshalPart(c.node.Part)
	if err != nil {
		return err // TODO: contextualise
	}
	props := &pb.NodeConfig{
		Name:         c.sharedOutlets.nodeNameInput.Get("value").String(),
		Enabled:      c.sharedOutlets.nodeEnabledInput.Get("checked").Bool(),
		Multiplicity: uint32(c.sharedOutlets.nodeMultiplicityInput.Get("value").Int()),
		Wait:         c.sharedOutlets.nodeWaitInput.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            c.node.X,
		Y:            c.node.Y,
	}
	req := &pb.SetNodePropertiesRequest{
		Graph: c.graph.FilePath,
		Node:  c.node.Name,
		Props: props,
	}
	if _, err := c.client.SetNodeProperties(ctx, req); err != nil {
		return err // TODO: contextualise
	}
	// Update local copy, since these were read at save time.
	// TODO: check whether the available pins have changed.
	if c.node.Name != props.Name {
		c.node.Name = props.Name
	}
	c.node.Enabled = props.Enabled
	c.node.Multiplicity = uint(props.Multiplicity)
	c.node.Wait = props.Wait
	return nil
}

func (c *nodeController) SetPosition(ctx context.Context, x, y float64) error {
	_, err := c.client.SetPosition(ctx, &pb.SetPositionRequest{
		Graph: c.graph.FilePath,
		Node:  c.node.Name,
		X:     x,
		Y:     y,
	})
	if err != nil {
		return err // TODO: contextualise
	}
	c.node.X, c.node.Y = x, y
	return nil
}

type focusable interface {
	GainFocus()
}

func (c *nodeController) GainFocus() {
	c.gc.showRHSPanel(c.gc.NodePropertiesPanel)

	c.sharedOutlets.nodeNameInput.Set("value", c.node.Name)
	c.sharedOutlets.nodeEnabledInput.Set("checked", c.node.Enabled)
	c.sharedOutlets.nodeMultiplicityInput.Set("value", c.node.Multiplicity)
	c.sharedOutlets.nodeWaitInput.Set("checked", c.node.Wait)
	c.sharedOutlets.nodePartEditors[c.node.Part.TypeKey()].Links.Show()
	if c.subpanel == nil {
		c.subpanel = c.sharedOutlets.nodeMetadataSubpanel
	}
	c.showSubpanel(c.subpanel)
	if f := c.node.Part.(focusable); f != nil {
		f.GainFocus()
	}
}

func (c *nodeController) LoseFocus() {
	// Nop.
}

func (c *nodeController) ShowMetadataSubpanel() {
	c.showSubpanel(c.sharedOutlets.nodeMetadataSubpanel)
}

func (c *nodeController) ShowPartSubpanel(name string) {
	c.showSubpanel(c.sharedOutlets.nodePartEditors[c.node.Part.TypeKey()].Panels[name])
}

func (c *nodeController) showSubpanel(p dom.Element) {
	c.subpanel = p
	if c.sharedOutlets.nodeCurrentSubpanel == p {
		return
	}
	c.sharedOutlets.nodeCurrentSubpanel.Hide()
	c.sharedOutlets.nodeCurrentSubpanel = p.Show()
}
