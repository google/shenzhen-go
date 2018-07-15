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
	"fmt"

	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

type partEditor struct {
	LinkGroup dom.Element
	Panels    map[string]*subpanel
}

type subpanel struct {
	Link  dom.Element
	Panel dom.Element
}

type nodeSharedOutlets struct {
	// Node properties subpanels and inputs
	subpanelMetadata  *subpanel
	subpanelCurrent   *subpanel
	inputName         dom.Element
	textareaComment   dom.Element
	inputEnabled      dom.Element
	inputMultiplicity dom.Element
	inputWait         dom.Element
	partEditors       map[string]*partEditor
}

type nodeController struct {
	client        pb.ShenzhenGoClient
	graph         *model.Graph
	node          *model.Node
	sharedOutlets *nodeSharedOutlets

	gc       *graphController
	subpanel *subpanel // remember most recent subpanel for each node
}

func (c *nodeController) Name() string {
	if c.node.Multiplicity == "1" {
		return c.node.Name
	}
	return fmt.Sprintf("%s (×%s)", c.node.Name, c.node.Multiplicity)
}

func (c *nodeController) Position() (x, y float64) { return c.node.X, c.node.Y }

func (c *nodeController) Pins(f func(view.PinController, string)) {
	pins := c.node.Pins()
	for name, conn := range c.node.Connections {
		pc := &pinController{
			client: c.client,
			graph:  c.graph,
			node:   c.node,
			name:   name,
			def:    pins[name],
		}
		f(pc, conn)
	}
}

func (c *nodeController) Delete(ctx context.Context) error {
	_, err := c.client.SetNode(ctx, &pb.SetNodeRequest{
		Graph: c.graph.FilePath,
		Node:  c.node.Name,
	})
	if err != nil {
		// TODO: contextualise
		return err
	}
	c.graph.DeleteNode(c.node, true)
	c.node = nil
	return nil
}

func (c *nodeController) Commit(ctx context.Context) error {
	if c.node == nil {
		return nil
	}
	pj, err := model.MarshalPart(c.node.Part)
	if err != nil {
		return err // TODO: contextualise
	}
	cfg := &pb.NodeConfig{
		Name:         c.sharedOutlets.inputName.Get("value").String(),
		Comment:      c.sharedOutlets.textareaComment.Get("value").String(),
		Enabled:      c.sharedOutlets.inputEnabled.Get("checked").Bool(),
		Multiplicity: c.sharedOutlets.inputMultiplicity.Get("value").String(),
		Wait:         c.sharedOutlets.inputWait.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            c.node.X,
		Y:            c.node.Y,
	}
	req := &pb.SetNodeRequest{
		Graph:  c.graph.FilePath,
		Node:   c.node.Name,
		Config: cfg,
	}
	if _, err := c.client.SetNode(ctx, req); err != nil {
		return err // TODO: contextualise
	}
	// Update local copy, since these were read at save time.
	c.graph.RenameNode(c.node, cfg.Name)
	c.node.Comment = cfg.Comment
	c.node.Enabled = cfg.Enabled
	c.node.Multiplicity = cfg.Multiplicity
	c.node.Wait = cfg.Wait
	c.node.RefreshConnections()
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
	c.gc.showRHSPanel(c.gc.nodePropertiesPanel)

	c.sharedOutlets.inputName.Set("value", c.node.Name)
	c.sharedOutlets.textareaComment.Set("value", c.node.Comment)
	c.sharedOutlets.inputEnabled.Set("checked", c.node.Enabled)
	c.sharedOutlets.inputMultiplicity.Set("value", c.node.Multiplicity)
	c.sharedOutlets.inputWait.Set("checked", c.node.Wait)
	// Hide all parteditor links except for this parttype
	for k, e := range c.sharedOutlets.partEditors {
		if k == c.node.Part.TypeKey() {
			e.LinkGroup.Show()
		} else {
			e.LinkGroup.Hide()
		}
	}
	c.showSubpanel(c.subpanel)
}

func (c *nodeController) ShowMetadataSubpanel() {
	c.showSubpanel(c.sharedOutlets.subpanelMetadata)
}

func (c *nodeController) ShowPartSubpanel(name string) {
	c.showSubpanel(c.sharedOutlets.partEditors[c.node.Part.TypeKey()].Panels[name])
}

func (c *nodeController) showSubpanel(p *subpanel) {
	if f, ok := c.node.Part.(focusable); ok {
		// Wait until after panel is shown in case of display weirdness.
		defer f.GainFocus()
	}

	c.subpanel = p
	if c.sharedOutlets.subpanelCurrent == p {
		return
	}
	c.sharedOutlets.subpanelCurrent.Link.ClassList().Remove("selected")
	c.sharedOutlets.subpanelCurrent.Panel.Hide()
	c.sharedOutlets.subpanelCurrent = p
	p.Link.ClassList().Add("selected")
	p.Panel.Show()
}
