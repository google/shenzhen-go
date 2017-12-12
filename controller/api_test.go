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
	"testing"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/parts"
	"github.com/google/shenzhen-go/model/pin"
	pb "github.com/google/shenzhen-go/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if st, ok := status.FromError(err); ok {
		return st.Code()
	}
	return codes.Unknown
}

func TestLookupGraph(t *testing.T) {
	foo := &model.Graph{Name: "foo"}
	bar := &model.Graph{Name: "bar"}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{
			"foo": foo,
			"bar": bar,
		},
	}
	tests := []struct {
		key  string
		g    *model.Graph
		code codes.Code
	}{
		{"foo", foo, codes.OK},
		{"bar", bar, codes.OK},
		{"baz", nil, codes.NotFound},
	}
	for _, test := range tests {
		g, err := c.lookupGraph(test.key)
		if got, want := g, test.g; got != want {
			t.Errorf("c.lookupGraph(%q) = %v, want %v", test.key, got, want)
		}
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.lookupGraph(%q) = %v, want %v", test.key, got, want)
		}
	}
}

func TestLookupNode(t *testing.T) {
	bar := &model.Node{Name: "bar"}
	foo := &model.Graph{
		Name:  "foo",
		Nodes: map[string]*model.Node{"bar": bar},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		gk, nk string
		g      *model.Graph
		n      *model.Node
		code   codes.Code
	}{
		{"foo", "bar", foo, bar, codes.OK},
		{"foo", "baz", nil, nil, codes.NotFound},
		{"baz", "bar", nil, nil, codes.NotFound},
	}
	for _, test := range tests {
		g, n, err := c.lookupNode(test.gk, test.nk)
		if got, want := g, test.g; got != want {
			t.Errorf("c.lookupNode(%q, %q) = graph %v, want %v", test.gk, test.nk, got, want)
		}
		if got, want := n, test.n; got != want {
			t.Errorf("c.lookupNode(%q, %q) = node %v, want %v", test.gk, test.nk, got, want)
		}
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.lookupNode(%q, %q) = code %v, want %v", test.gk, test.nk, got, want)
		}
	}
}

func TestLookupChannel(t *testing.T) {
	bar := &model.Channel{Name: "bar"}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		gk, ck string
		g      *model.Graph
		ch     *model.Channel
		code   codes.Code
	}{
		{"foo", "bar", foo, bar, codes.OK},
		{"foo", "baz", nil, nil, codes.NotFound},
		{"baz", "bar", nil, nil, codes.NotFound},
	}
	for _, test := range tests {
		g, ch, err := c.lookupChannel(test.gk, test.ck)
		if got, want := g, test.g; got != want {
			t.Errorf("c.lookupChannel(%q, %q) = graph %v, want %v", test.gk, test.ck, got, want)
		}
		if got, want := ch, test.ch; got != want {
			t.Errorf("c.lookupChannel(%q, %q) = node %v, want %v", test.gk, test.ck, got, want)
		}
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.lookupChannel(%q, %q) = code %v, want %v", test.gk, test.ck, got, want)
		}
	}
}

func TestCreateChannel(t *testing.T) {
	bar := &model.Channel{Name: "bar"}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.CreateChannelRequest
		code codes.Code
	}{
		{
			req: &pb.CreateChannelRequest{
				Graph: "nope",
				Name:  "baz",
			},
			code: codes.NotFound,
		},
		{
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
			},
			code: codes.OK,
		},
		{
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "bar",
			},
			code: codes.FailedPrecondition,
		},
	}
	for _, test := range tests {
		_, err := c.CreateChannel(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.CreateChannel(%v) = code %v, want %v", test.req, got, want)
		}
	}
	_, got := foo.Channels["baz"]
	if want := true; got != want {
		t.Errorf("foo.Channels[baz] is missing, want present")
	}
}

func TestConnectPin(t *testing.T) {
	baz := &model.Node{
		Name: "baz",
		Part: parts.NewCode(nil, "", "", "", pin.Map{
			"qux": &pin.Definition{
				Type: "int",
			},
		}),
		Connections: make(map[string]string),
	}
	bar := &model.Channel{Name: "bar", Type: "int"}
	tuz := &model.Channel{Name: "tuz", Type: "string"}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar, "tuz": tuz},
		Nodes:    map[string]*model.Node{"baz": baz},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.ConnectPinRequest
		code codes.Code
	}{
		{
			// no such graph
			req: &pb.ConnectPinRequest{
				Graph:   "nope",
				Node:    "baz",
				Channel: "bar",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			// no such node
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "barz",
				Channel: "bar",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			// no such channel
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "buz",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			// No such pin
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "bar",
				Pin:     "wux",
			},
			code: codes.NotFound,
		},
		{
			// type mismatch
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "tuz",
				Pin:     "qux",
			},
			code: codes.FailedPrecondition,
		},
		{
			// It works
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "bar",
				Pin:     "qux",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.ConnectPin(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.ConnectPin(%v) = code %v, want %v", test.req, got, want)
		}
	}
	if got, want := baz.Connections["qux"], "bar"; got != want {
		t.Errorf("baz.Connections[qux] = %q, want %q", got, want)
	}
}

func TestDeleteChannel(t *testing.T) {
	baz := &model.Node{
		Name: "baz",
		Part: parts.NewCode(nil, "", "", "", pin.Map{
			"qux": &pin.Definition{
				Type: "int",
			},
		}),
		Connections: map[string]string{
			"qux": "bar",
		},
	}
	bar := &model.Channel{
		Name: "bar",
		Type: "int",
		Pins: map[model.NodePin]struct{}{
			{Node: "baz", Pin: "qux"}: {},
		},
	}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar},
		Nodes:    map[string]*model.Node{"baz": baz},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.DeleteChannelRequest
		code codes.Code
	}{
		{ // No such graph
			req: &pb.DeleteChannelRequest{
				Graph:   "nope",
				Channel: "bar",
			},
			code: codes.NotFound,
		},
		{ // No such channel
			req: &pb.DeleteChannelRequest{
				Graph:   "foo",
				Channel: "baz",
			},
			code: codes.NotFound,
		},
		{ // Ok
			req: &pb.DeleteChannelRequest{
				Graph:   "foo",
				Channel: "bar",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.DeleteChannel(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.DeleteChannel(%v) = code %v, want %v", test.req, got, want)
		}
	}
	// Channel should be gone
	if _, found := foo.Channels["bar"]; found {
		t.Error("channel 'bar' still exists in graph")
	}
	// Reference from node should be gone
	if got, want := baz.Connections["qux"], "nil"; got != want {
		t.Errorf("baz.Connections[qux] = %q, want %q", got, want)
	}
}

func TestDisconnectPin(t *testing.T) {
	baz := &model.Node{
		Name: "baz",
		Part: parts.NewCode(nil, "", "", "", pin.Map{
			"qux": &pin.Definition{
				Type: "int",
			},
		}),
		Connections: map[string]string{
			"qux": "bar",
		},
	}
	bar := &model.Channel{
		Name: "bar",
		Type: "int",
		Pins: map[model.NodePin]struct{}{
			{Node: "baz", Pin: "qux"}: {},
		},
	}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar},
		Nodes:    map[string]*model.Node{"baz": baz},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.DisconnectPinRequest
		code codes.Code
	}{
		{ // No such graph
			req: &pb.DisconnectPinRequest{
				Graph: "nope",
				Node:  "baz",
				Pin:   "qux",
			},
			code: codes.NotFound,
		},
		{ // No such node
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "bar",
				Pin:   "qux",
			},
			code: codes.NotFound,
		},
		{ // No such pin
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "baz",
				Pin:   "tuz",
			},
			code: codes.NotFound,
		},
		{ // Ok
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "baz",
				Pin:   "qux",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.DisconnectPin(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.DisconnectPin(%v) = code %v, want %v", test.req, got, want)
		}
	}
	// Reference from node should be gone
	if got, want := baz.Connections["qux"], "nil"; got != want {
		t.Errorf("baz.Connections[qux] = %q, want %q", got, want)
	}
	// Channel should be gone
	if _, found := foo.Channels["bar"]; found {
		t.Error("channel 'bar' still exists in graph")
	}
}

func TestSave(t *testing.T) {
	// TODO
}

func TestSetGraphProperties(t *testing.T) {
	foo := &model.Graph{Name: "foo"}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{
			"foo": foo,
		},
	}
	tests := []struct {
		req  *pb.SetGraphPropertiesRequest
		code codes.Code
	}{
		{
			req: &pb.SetGraphPropertiesRequest{
				Graph: "oof",
			},
			code: codes.NotFound,
		},
		{
			req: &pb.SetGraphPropertiesRequest{
				Graph:       "foo",
				Name:        "name",
				PackagePath: "package/path",
				IsCommand:   true,
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.SetGraphProperties(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.SetGraphProperties(%v) = code %v, want %v", test.req, got, want)
		}
	}
	if got, want := foo.Name, "name"; got != want {
		t.Errorf("foo.Name = %q, want %q", got, want)
	}
	if got, want := foo.PackagePath, "package/path"; got != want {
		t.Errorf("foo.PackagePath = %q, want %q", got, want)
	}
	if got, want := foo.IsCommand, true; got != want {
		t.Errorf("foo.IsCommand = %t, want %t", got, want)
	}
}

func TestSetNodeProperties(t *testing.T) {
	bar := &model.Node{Name: "bar"}
	baz := &model.Node{Name: "baz"}
	foo := &model.Graph{
		Name: "foo",
		Nodes: map[string]*model.Node{
			"bar": bar,
			"baz": baz,
		},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.SetNodePropertiesRequest
		code codes.Code
	}{
		{ // no such graph
			req: &pb.SetNodePropertiesRequest{
				Graph: "nope",
				Node:  "bar",
			},
			code: codes.NotFound,
		},
		{ // no such node
			req: &pb.SetNodePropertiesRequest{
				Graph: "foo",
				Node:  "bak",
			},
			code: codes.NotFound,
		},
		{ // can't unmarshal
			req: &pb.SetNodePropertiesRequest{
				Graph:    "foo",
				Node:     "bar",
				PartType: "Not a part key",
			},
			code: codes.FailedPrecondition,
		},
		{ // rename to existing name
			req: &pb.SetNodePropertiesRequest{
				Graph:    "foo",
				Node:     "bar",
				Name:     "baz",
				PartCfg:  []byte("{}"),
				PartType: "Code",
			},
			code: codes.FailedPrecondition,
		},
		{ // Ok
			req: &pb.SetNodePropertiesRequest{
				Graph:        "foo",
				Node:         "bar",
				Name:         "bax",
				PartCfg:      []byte("{}"),
				PartType:     "Code",
				Multiplicity: 1,
				Enabled:      true,
				Wait:         true,
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.SetNodeProperties(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.SetNodeProperties(%v) = code %v, want %v", test.req, got, want)
		}
	}
	if _, found := foo.Nodes["bar"]; found {
		t.Error("foo.Nodes[bar] is found, want not found")
	}
	bax := bar
	if got, want := foo.Nodes["bax"], bax; got != want {
		t.Errorf("foo.Nodes[bax] = %v, want %v", got, want)
	}
	if got, want := bax.Name, "bax"; got != want {
		t.Errorf("bax.Name = %q, want %q", got, want)
	}
	if got, want := bax.Multiplicity, uint(1); got != want {
		t.Errorf("bax.Multiplicity = %v, want %v", got, want)
	}
	if got, want := bax.Enabled, true; got != want {
		t.Errorf("bax.Enabled = %t, want %t", got, want)
	}
	if got, want := bax.Wait, true; got != want {
		t.Errorf("bax.Wait = %t, want %t", got, want)
	}
	if got, want := bax.Part.TypeKey(), "Code"; got != want {
		t.Errorf("bax.Part.TypeKey() = %q, want %q", got, want)
	}
}

func TestSetPosition(t *testing.T) {
	bar := &model.Node{Name: "bar"}
	foo := &model.Graph{
		Name:  "foo",
		Nodes: map[string]*model.Node{"bar": bar},
	}
	c := &controller{
		loadedGraphs: map[string]*model.Graph{"foo": foo},
	}
	tests := []struct {
		req  *pb.SetPositionRequest
		code codes.Code
	}{
		{
			req: &pb.SetPositionRequest{
				Graph: "nope",
				Node:  "bar",
			},
			code: codes.NotFound,
		},
		{
			req: &pb.SetPositionRequest{
				Graph: "foo",
				Node:  "baz",
			},
			code: codes.NotFound,
		},
		{
			req: &pb.SetPositionRequest{
				Graph: "foo",
				Node:  "bar",
				X:     42,
				Y:     17,
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		_, err := c.SetPosition(context.Background(), test.req)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.SetPosition(%v) = code %v, want %v", test.req, got, want)
		}
	}
	if got, want := bar.X, 42; got != want {
		t.Errorf("bar.X = %d, want %d", got, want)
	}
	if got, want := bar.Y, 17; got != want {
		t.Errorf("bar.Y = %d, want %d", got, want)
	}
}
