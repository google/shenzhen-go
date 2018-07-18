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
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
	"github.com/gopherjs/gopherjs/js"
)

const defaultChannelNamePrefix = "channel"

var (
	ace   = dom.GlobalAce()
	hterm = dom.GlobalHterm()

	aceTheme = dom.Global("aceTheme").String()

	_ view.GraphController = (*graphController)(nil)
)

type graphController struct {
	doc    dom.Document
	graph  *model.Graph
	client pb.ShenzhenGoClient

	// RHS panels
	currentRHSPanel        dom.Element
	channelPropertiesPanel dom.Element
	graphPropertiesPanel   dom.Element
	helpAboutPanel         dom.Element
	helpLicensesPanel      dom.Element
	htermPanel             dom.Element
	htermContainer         dom.Element
	htermTerminal          dom.Terminal
	nodePropertiesPanel    dom.Element
	previewGoPanel         dom.Element
	previewJSONPanel       dom.Element
	previewGoSession       *dom.AceSession
	previewJSONSession     *dom.AceSession

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
	e.SetTheme("ace/theme/" + aceTheme)
	return e.Session().SetMode(mode)
}

// NewGraphController returns a new controller for a graph, and binds outlets.
func NewGraphController(doc dom.Document, graph *model.Graph, client pb.ShenzhenGoClient) view.GraphController {
	pes := make(map[string]*partEditor, len(model.PartTypes))
	for n, t := range model.PartTypes {
		p := make(map[string]*subpanel, len(t.Panels))
		for _, d := range t.Panels {
			p[d.Name] = &subpanel{
				Link:  doc.ElementByID("node-" + n + "-" + d.Name + "-link"),
				Panel: doc.ElementByID("node-" + n + "-" + d.Name + "-panel"),
			}
		}
		pes[n] = &partEditor{
			LinkGroup: doc.ElementByID("node-" + n + "-links"),
			Panels:    p,
		}
	}

	subpanelMetadata := &subpanel{
		Link:  doc.ElementByID("node-metadata-link"),
		Panel: doc.ElementByID("node-metadata-panel"),
	}

	return &graphController{
		doc:    doc,
		client: client,
		graph:  graph,

		currentRHSPanel: doc.ElementByID("graph-properties"),

		channelPropertiesPanel: doc.ElementByID("channel-properties"),
		graphPropertiesPanel:   doc.ElementByID("graph-properties"),
		helpAboutPanel:         doc.ElementByID("help-about-panel"),
		helpLicensesPanel:      doc.ElementByID("help-licenses-panel"),
		htermPanel:             doc.ElementByID("hterm-panel"),
		htermContainer:         doc.ElementByID("hterm-terminal"),
		nodePropertiesPanel:    doc.ElementByID("node-properties"),
		previewGoPanel:         doc.ElementByID("preview-go"),
		previewJSONPanel:       doc.ElementByID("preview-json"),
		previewGoSession:       setupAceView("preview-go-ace", dom.AceGoMode),
		previewJSONSession:     setupAceView("preview-json-ace", dom.AceJSONMode),

		graphNameTextInput:        doc.ElementByID("graph-prop-name"),
		graphPackagePathTextInput: doc.ElementByID("graph-prop-package-path"),
		graphIsCommandCheckbox:    doc.ElementByID("graph-prop-is-command"),

		channelSharedOutlets: &channelSharedOutlets{
			inputName:     doc.ElementByID("channel-name"),
			codeType:      doc.ElementByID("channel-type"),
			inputCapacity: doc.ElementByID("channel-capacity"),
		},
		nodeSharedOutlets: &nodeSharedOutlets{
			subpanelMetadata:  subpanelMetadata,
			subpanelCurrent:   subpanelMetadata,
			inputName:         doc.ElementByID("node-name"),
			textareaComment:   doc.ElementByID("node-comment"),
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
	c.showRHSPanel(c.graphPropertiesPanel)
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

	for _, pc := range pcs {
		ch.AddPin(pc.NodeName(), pc.Name())
	}

	// Pick a unique name
	// River gently flows in spring
	// First try channelN
	ch.Name = defaultChannelNamePrefix + strconv.Itoa(len(c.graph.Channels))
	if c.graph.Channels[ch.Name] != nil {
		// Fall back to probing...
		for i := 0; ; i++ {
			ch.Name = defaultChannelNamePrefix + strconv.Itoa(i)
			if c.graph.Channels[ch.Name] == nil {
				break
			}
		}
	}

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
		Multiplicity: "1",
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
			Multiplicity: n.Multiplicity,
			PartType:     partType,
			PartCfg:      pm.Part,
		},
	})
	if err != nil {
		return nil, err
	}
	c.graph.Nodes[n.Name] = n
	n.RefreshConnections()
	return c.newNodeController(n), nil
}

func (c *graphController) action(ctx context.Context, a pb.ActionRequest_Action) error {
	stream, err := c.client.Action(ctx, &pb.ActionRequest{
		Graph:  c.graph.FilePath,
		Action: a,
	})
	if err != nil {
		return err
	}
	if a == pb.ActionRequest_SAVE || a == pb.ActionRequest_REVERT {
		// No need for a terminal
		return nil
	}
	c.ShowHterm()
	c.htermTerminal.ClearHome()
	tio := c.htermTerminal.IO().Push()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		tio.Print(resp.Output)
	}
	return nil
}

func (c *graphController) Save(ctx context.Context) error {
	return c.action(ctx, pb.ActionRequest_SAVE)
}

func (c *graphController) Revert(ctx context.Context) error {
	if err := c.action(ctx, pb.ActionRequest_REVERT); err != nil {
		return err
	}
	// TODO: Less janky reloading. (call into view reload)
	js.Global.Get("location").Call("reload", true)
	return nil
}

func (c *graphController) Generate(ctx context.Context) error {
	return c.action(ctx, pb.ActionRequest_GENERATE)
}

func (c *graphController) Build(ctx context.Context) error {
	return c.action(ctx, pb.ActionRequest_BUILD)
}

func (c *graphController) Install(ctx context.Context) error {
	return c.action(ctx, pb.ActionRequest_INSTALL)
}

func setupHterm(el dom.Element) dom.Terminal {
	wait := make(chan struct{})
	t := hterm.NewTerminal("default")
	t.OnTerminalReady(func() { close(wait) })
	t.SetAutoCR(true)
	t.Decorate(el)
	t.InstallKeyboard()
	<-wait
	return t
}

func (c *graphController) ShowHterm() {
	c.showRHSPanel(c.htermPanel)
	if c.htermTerminal.Object == nil {
		c.htermTerminal = setupHterm(c.htermContainer)
	}
}

func (c *graphController) Run(ctx context.Context) error {
	// TODO: Implement ^C / stop

	c.ShowHterm()
	c.htermTerminal.ClearHome()

	rc, err := c.client.Run(ctx)
	if err != nil {
		return err
	}
	defer rc.CloseSend()
	if err := rc.Send(&pb.Input{Graph: c.graph.FilePath}); err != nil {
		return err
	}

	tio := c.htermTerminal.IO().Push()
	defer func() {
		tio.OnVTKeystroke(func(string) {})
		tio.SendString(func(string) {})
		tio.Pop()
	}()

	var buf []byte
	tio.OnVTKeystroke(func(s string) {
		switch s {
		case "\r":
			rc.Send(&pb.Input{In: string(buf) + "\n"})
			buf = buf[:0]
			tio.Print("\n")
		case "\b", "\x7f":
			if len(buf) == 0 {
				return
			}
			buf = buf[:len(buf)-1]
			// I have no idea what I'm doing, do I
			tio.Print("\b \b")
		default:
			buf = append(buf, s...)
			tio.Print(s)
		}
	})
	tio.SendString(func(s string) {
		rc.Send(&pb.Input{In: string(buf) + s})
		buf = buf[:0]
		tio.Print(s)
	})

	for {
		out, err := rc.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// TODO(josh): Format these differently?
		tio.Print(out.Out)
		tio.Print(out.Err)
	}
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

func (c *graphController) PreviewRawGo() {
	g, err := c.graph.RawGo()
	if err != nil {
		// TODO: better place to put output?
		g = "/* Couldn't generate raw Go: \n" + err.Error() + " */"
	}
	c.previewGoSession.SetValue(g)
	c.showRHSPanel(c.previewGoPanel)
}

func (c *graphController) PreviewGo() {
	g, err := c.graph.Go()
	if err != nil {
		// TODO: better place to put output?
		g = "/* Couldn't generate Go: \n" + err.Error() + " */"
	}
	c.previewGoSession.SetValue(g)
	c.showRHSPanel(c.previewGoPanel)
}

func (c *graphController) PreviewJSON() {
	g, err := c.graph.JSON()
	if err != nil {
		// TODO: better place to put output?
		g = `{"error": "Couldn't generate JSON: ` + err.Error() + `"}`
	}
	c.previewJSONSession.SetValue(g)
	c.showRHSPanel(c.previewJSONPanel)
}

func (c *graphController) HelpLicenses() { c.showRHSPanel(c.helpLicensesPanel) }
func (c *graphController) HelpAbout()    { c.showRHSPanel(c.helpAboutPanel) }

// showRHSPanel hides any existing panel and shows the given element as the panel.
func (c *graphController) showRHSPanel(p dom.Element) {
	if p == c.currentRHSPanel {
		return
	}
	c.currentRHSPanel.Hide()
	c.currentRHSPanel = p.Show()
}
