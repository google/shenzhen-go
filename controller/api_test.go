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
	// TODO
}

func TestDisconnectPin(t *testing.T) {
	// TODO
}

func TestSave(t *testing.T) {
	// TODO
}

func TestSetGraphProperties(t *testing.T) {
	// TODO
}

func TestSetNodeProperties(t *testing.T) {
	// TODO
}

func TestSetPosition(t *testing.T) {
	// TODO
}
