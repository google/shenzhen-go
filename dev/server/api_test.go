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
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/parts"
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

func TestSetChannelCreate(t *testing.T) {
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
		req  *pb.SetChannelRequest
		code codes.Code
	}{
		{
			name: "graph 'nope' doesn't exist",
			req: &pb.SetChannelRequest{
				Graph:  "nope",
				Config: &pb.ChannelConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "channel 'bar' already exists",
			req: &pb.SetChannelRequest{
				Graph: "foo",
				Config: &pb.ChannelConfig{
					Name: "bar",
				},
			},
			code: codes.AlreadyExists,
		},
		{
			name: "node 'nope' doesn't exist",
			req: &pb.SetChannelRequest{
				Graph: "foo",
				Config: &pb.ChannelConfig{
					Name: "baz",
					Pins: []*pb.NodePin{
						{Node: "nope", Pin: "pin1"},
					},
				},
			},
			code: codes.NotFound,
		},
		{
			name: "pin 'pine' doesn't exist",
			req: &pb.SetChannelRequest{
				Graph: "foo",
				Config: &pb.ChannelConfig{
					Name: "baz",
					Pins: []*pb.NodePin{
						{Node: "node1", Pin: "pine"},
					},
				},
			},
			code: codes.NotFound,
		},
		{
			name: "ok",
			req: &pb.SetChannelRequest{
				Graph: "foo",
				Config: &pb.ChannelConfig{
					Name: "baz",
					Pins: []*pb.NodePin{
						{Node: "node1", Pin: "pin1"},
						{Node: "node2", Pin: "pin2"},
					},
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetChannel(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Fatalf("c.SetChannel(%v) = error %v, want %v", test.req, err, want)
			}
			wantBaz, wantCon := true, "baz"
			if err != nil {
				wantBaz, wantCon = false, "nil"
			}
			_, got := foo.Channels["baz"]
			if want := wantBaz; got != want {
				t.Errorf("after c.SetChannel(%v): foo.Channels[baz] = _, %t, want %t", test.req, got, want)
			}
			if got, want := node1.Connections["pin1"], wantCon; got != want {
				t.Errorf("after c.SetChannel(%v): node1.Connections[pin1] = %q, want %q", test.req, got, want)
			}
			if got, want := node2.Connections["pin2"], wantCon; got != want {
				t.Errorf("after c.SetChannel(%v): node2.Connections[pin2] = %q, want %q", test.req, got, want)
			}
		})
	}
}

func TestSetNodeCreate(t *testing.T) {
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
		req  *pb.SetNodeRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.SetNodeRequest{
				Graph:  "nope",
				Config: &pb.NodeConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "can't unmarshal",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Config: &pb.NodeConfig{
					PartType: "Not a part key",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "existing name",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Config: &pb.NodeConfig{
					Name:     "baz",
					PartCfg:  []byte("{}"),
					PartType: "Code",
				},
			},
			code: codes.AlreadyExists,
		},
		{
			name: "Ok",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Config: &pb.NodeConfig{
					Name:         "bax",
					PartCfg:      []byte("{}"),
					PartType:     "Code",
					Multiplicity: "1",
					Enabled:      true,
					Wait:         true,
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetNode(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetNode(%v) = error %v, want %v", test.req, err, want)
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
	if got, want := bax.Multiplicity, "1"; got != want {
		t.Errorf("bax.Multiplicity = %q, want %q", got, want)
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

func TestSetChannelDelete(t *testing.T) {
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
		req  *pb.SetChannelRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.SetChannelRequest{
				Graph:   "nope",
				Channel: "bar",
				// Config: nil,
			},
			code: codes.NotFound,
		},
		{
			name: "No such channel",
			req: &pb.SetChannelRequest{
				Graph:   "foo",
				Channel: "baz",
				// Config: nil,
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
			req: &pb.SetChannelRequest{
				Graph:   "foo",
				Channel: "bar",
				// Config: nil,
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetChannel(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetChannel(%v) = code %v, want %v", test.req, got, want)
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

func TestSetNodeDelete(t *testing.T) {
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
		req  *pb.SetNodeRequest
		code codes.Code
	}{
		{
			name: "No such graph",
			req: &pb.SetNodeRequest{
				Graph: "nope",
				Node:  "baz",
				// Config: nil,
			},
			code: codes.NotFound,
		},
		{
			name: "No such channel",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "bar",
				// Config: nil,
			},
			code: codes.NotFound,
		},
		{
			name: "Ok",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "baz",
				// Config: nil,
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetNode(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetNode(%v) = code %v, want %v", test.req, got, want)
			}
		})
	}
	// Node should be gone
	if _, found := foo.Nodes["baz"]; found {
		t.Error("node 'baz' still exists in graph")
	}
	// Reference from channel should be gone
	if bar.HasPin("baz", "qux") {
		t.Error("channel 'bar' still references node 'baz' pin 'qux'")
	}
}

func TestSave(t *testing.T) {
	// TODO: implement tests for saving
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

func TestSetNode(t *testing.T) {
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
		req  *pb.SetNodeRequest
		code codes.Code
	}{
		{
			name: "no such graph",
			req: &pb.SetNodeRequest{
				Graph:  "nope",
				Node:   "bar",
				Config: &pb.NodeConfig{},
			},
			code: codes.NotFound,
		},
		{
			name: "no such node",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "bak",
				Config: &pb.NodeConfig{
					PartCfg:  []byte("{}"),
					PartType: "Code",
				},
			},
			code: codes.NotFound,
		},
		{
			name: "can't unmarshal part",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "bar",
				Config: &pb.NodeConfig{
					PartType: "Not a part key",
				},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "rename to existing name",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "bar",
				Config: &pb.NodeConfig{
					Name:     "baz",
					PartCfg:  []byte("{}"),
					PartType: "Code",
				},
			},
			code: codes.AlreadyExists,
		},
		{
			name: "Ok",
			req: &pb.SetNodeRequest{
				Graph: "foo",
				Node:  "bar",
				Config: &pb.NodeConfig{
					Name:         "bax",
					PartCfg:      []byte("{}"),
					PartType:     "Code",
					Multiplicity: "1",
					Enabled:      true,
					Wait:         true,
				},
			},
			code: codes.OK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.SetNode(context.Background(), test.req)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.SetNodeProperties(%v) = error %v, want code %v", test.req, err, want)
			}
		})
	}
	if _, found := foo.Nodes["bar"]; found {
		t.Error("foo.Nodes[bar] is found, want not found")
	}
	bax := foo.Nodes["bax"]
	if bax == nil {
		t.Error("foo.Nodes[bax] = nil, want new node")
	}
	if got, want := bax.Name, "bax"; got != want {
		t.Errorf("bax.Name = %q, want %q", got, want)
	}
	if got, want := bax.Multiplicity, "1"; got != want {
		t.Errorf("bax.Multiplicity = %q, want %q", got, want)
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
