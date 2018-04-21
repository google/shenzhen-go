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

package server

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/parts"
	"github.com/google/shenzhen-go/dev/model/pin"
	pb "github.com/google/shenzhen-go/dev/proto/go"
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

func TestCreateChannel(t *testing.T) {
	bar := &model.Channel{Name: "bar"}
	node1 := &model.Node{
		Name: "node1",
		Connections: map[string]string{
			"pin1": "nil",
		}}
	node2 := &model.Node{
		Name: "node2",
		Connections: map[string]string{
			"pin2": "nil",
		}}
	foo := &model.Graph{
		Name: "foo",
		Channels: map[string]*model.Channel{
			"bar": bar,
		},
		Nodes: map[string]*model.Node{
			"node1": node1,
			"node2": node2,
		},
	}
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.CreateChannelRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.CreateChannelRequest{
				Graph: "nope",
				Name:  "baz",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pin1",
					},
					{
						Node: "node2",
						Pin:  "pin2",
					},
				},
			},
			code: codes.NotFound,
		},
		{
			name: "channel already exists",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "bar",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pin1",
					},
					{
						Node: "node2",
						Pin:  "pin2",
					},
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "node1 doesn't exist",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
				Pins: []*pb.NodePin{
					{
						Node: "nope",
						Pin:  "pin1",
					},
					{
						Node: "node2",
						Pin:  "pin2",
					},
				},
			},
			code: codes.NotFound,
		},
		{
			name: "node2 doesn't exist",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pin1",
					},
					{
						Node: "noop",
						Pin:  "pin2",
					},
				},
			},
			code: codes.NotFound,
		},
		{
			name: "pin1 doesn't exist",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pine",
					},
					{
						Node: "node2",
						Pin:  "pin2",
					},
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "pin2 doesn't exist",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pin1",
					},
					{
						Node: "node2",
						Pin:  "pint",
					},
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "ok",
			req: &pb.CreateChannelRequest{
				Graph: "foo",
				Name:  "baz",
				Type:  "int",
				Pins: []*pb.NodePin{
					{
						Node: "node1",
						Pin:  "pin1",
					},
					{
						Node: "node2",
						Pin:  "pin2",
					},
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.CreateChannel(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Fatalf("c.CreateChannel(%v) = error %v, want %v", test.req, err, want)
			}
			wantBaz, wantCon := true, "baz"
			if err != nil {
				wantBaz, wantCon = false, "nil"
			}
			_, got := foo.Channels["baz"]
			if want := wantBaz; got != want {
				t.Errorf("after c.CreateChannel(%v): foo.Channels[baz] = _, %t, want %t", test.req, got, want)
			}
			if got, want := node1.Connections["pin1"], wantCon; got != want {
				t.Errorf("after c.CreateChannel(%v): node1.Connections[pin1] = %q, want %q", test.req, got, want)
			}
			if got, want := node2.Connections["pin2"], wantCon; got != want {
				t.Errorf("after c.CreateChannel(%v): node2.Connections[pin2] = %q, want %q", test.req, got, want)
			}
		})
	}
}

func TestCreateNode(t *testing.T) {
	baz := &model.Node{Name: "baz"}
	foo := &model.Graph{
		Name: "foo",
		Nodes: map[string]*model.Node{
			"baz": baz,
		},
	}
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.CreateNodeRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.CreateNodeRequest{
				Graph: "nope",
				Props: &pb.NodeConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "can't unmarshal",
			req: &pb.CreateNodeRequest{
				Graph: "foo",
				Props: &pb.NodeConfig{
					PartType: "Not a part key",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "existing name",
			req: &pb.CreateNodeRequest{
				Graph: "foo",
				Props: &pb.NodeConfig{
					Name:     "baz",
					PartCfg:  []byte("{}"),
					PartType: "Code",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "Ok",
			req: &pb.CreateNodeRequest{
				Graph: "foo",
				Props: &pb.NodeConfig{
					Name:         "bax",
					PartCfg:      []byte("{}"),
					PartType:     "Code",
					Multiplicity: 1,
					Enabled:      true,
					Wait:         true,
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.CreateNode(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.CreateNode(%v) = error %v, want %v", test.req, err, want)
			}
		})
	}
	bax := foo.Nodes["bax"]
	if bax == nil {
		t.Error("foo.Nodes[bax] = nil, want bax")
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
	bar := &model.Channel{
		Name: "bar",
		Type: "int",
		Pins: make(map[model.NodePin]struct{}),
	}
	tuz := &model.Channel{
		Name: "tuz",
		Type: "string",
		Pins: make(map[model.NodePin]struct{}),
	}
	foo := &model.Graph{
		Name:     "foo",
		Channels: map[string]*model.Channel{"bar": bar, "tuz": tuz},
		Nodes:    map[string]*model.Node{"baz": baz},
	}
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.ConnectPinRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.ConnectPinRequest{
				Graph:   "nope",
				Node:    "baz",
				Channel: "bar",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			name: "no such node",
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "barz",
				Channel: "bar",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			name: "no such channel",
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "buz",
				Pin:     "qux",
			},
			code: codes.NotFound,
		},
		{
			name: "No such pin",
			req: &pb.ConnectPinRequest{
				Graph:   "foo",
				Node:    "baz",
				Channel: "bar",
				Pin:     "wux",
			},
			code: codes.NotFound,
		},
		{
			name: "It works",
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
		t.Run(test.name, func(t *testing.T) {
			_, err := c.ConnectPin(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.ConnectPin(%v) = code %v, want %v", test.req, got, want)
			}
		})
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.DeleteChannelRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.DeleteChannelRequest{
				Graph:   "nope",
				Channel: "bar",
			},
			code: codes.NotFound,
		},
		{
			name: "No such channel",
			req: &pb.DeleteChannelRequest{
				Graph:   "foo",
				Channel: "baz",
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
			req: &pb.DeleteChannelRequest{
				Graph:   "foo",
				Channel: "bar",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.DeleteChannel(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.DeleteChannel(%v) = code %v, want %v", test.req, got, want)
			}
		})
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

func TestDeleteNode(t *testing.T) {
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.DeleteNodeRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.DeleteNodeRequest{
				Graph: "nope",
				Node:  "baz",
			},
			code: codes.NotFound,
		},
		{
			name: "No such channel",
			req: &pb.DeleteNodeRequest{
				Graph: "foo",
				Node:  "bar",
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
			req: &pb.DeleteNodeRequest{
				Graph: "foo",
				Node:  "baz",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.DeleteNode(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.DeleteNode(%v) = code %v, want %v", test.req, got, want)
			}
		})
	}
	// Channel should be gone
	if _, found := foo.Channels["bar"]; found {
		t.Error("channel 'bar' still exists in graph")
	}
	// Node should be gone
	if _, found := foo.Nodes["baz"]; found {
		t.Error("node 'baz' still exists in graph")
	}
	// Reference from node should be gone
	if got, want := baz.Connections["qux"], "nil"; got != want {
		t.Errorf("baz.Connections[qux] = %q, want %q", got, want)
	}
	// Reference from channel should be gone
	if _, found := bar.Pins[model.NodePin{Node: "baz", Pin: "qux"}]; found {
		t.Error("channel 'bar' still references node 'baz' pin 'qux'")
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.DisconnectPinRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.DisconnectPinRequest{
				Graph: "nope",
				Node:  "baz",
				Pin:   "qux",
			},
			code: codes.NotFound,
		},
		{
			name: "No such node",
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "bar",
				Pin:   "qux",
			},
			code: codes.NotFound,
		},
		{
			name: "No such pin",
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "baz",
				Pin:   "tuz",
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
			req: &pb.DisconnectPinRequest{
				Graph: "foo",
				Node:  "baz",
				Pin:   "qux",
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.DisconnectPin(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.DisconnectPin(%v) = code %v, want %v", test.req, got, want)
			}
		})
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{
			"foo": {Graph: foo},
		},
	}
	tests := []struct {
		name string
		req  *pb.SetGraphPropertiesRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.SetGraphPropertiesRequest{
				Graph: "oof",
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
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
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetGraphProperties(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetGraphProperties(%v) = code %v, want %v", test.req, got, want)
			}
		})
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.SetNodePropertiesRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.SetNodePropertiesRequest{
				Graph: "nope",
				Node:  "bar",
				Props: &pb.NodeConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "no such node",
			req: &pb.SetNodePropertiesRequest{
				Graph: "foo",
				Node:  "bak",
				Props: &pb.NodeConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "can't unmarshal",
			req: &pb.SetNodePropertiesRequest{
				Graph: "foo",
				Node:  "bar",
				Props: &pb.NodeConfig{
					PartType: "Not a part key",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "rename to existing name",
			req: &pb.SetNodePropertiesRequest{
				Graph: "foo",
				Node:  "bar",
				Props: &pb.NodeConfig{
					Name:     "baz",
					PartCfg:  []byte("{}"),
					PartType: "Code",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "Ok",
			req: &pb.SetNodePropertiesRequest{
				Graph: "foo",
				Node:  "bar",
				Props: &pb.NodeConfig{
					Name:         "bax",
					PartCfg:      []byte("{}"),
					PartType:     "Code",
					Multiplicity: 1,
					Enabled:      true,
					Wait:         true,
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetNodeProperties(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetNodeProperties(%v) = code %v, want %v", test.req, got, want)
			}
		})
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
	c := &server{
		loadedGraphs: map[string]*serveGraph{"foo": {Graph: foo}},
	}
	tests := []struct {
		name string
		req  *pb.SetPositionRequest
		code codes.Code
	}{
		{
			name: "Graph not found",
			req: &pb.SetPositionRequest{
				Graph: "nope",
				Node:  "bar",
			},
			code: codes.NotFound,
		},
		{
			name: "Node not found",
			req: &pb.SetPositionRequest{
				Graph: "foo",
				Node:  "baz",
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
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
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetPosition(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetPosition(%v) = code %v, want %v", test.req, got, want)
			}
		})
	}
	if got, want := bar.X, 42.; got != want {
		t.Errorf("bar.X = %f, want %f", got, want)
	}
	if got, want := bar.Y, 17.; got != want {
		t.Errorf("bar.Y = %f, want %f", got, want)
	}
}
