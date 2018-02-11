package server

import (
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/google/shenzhen-go/model"
)

func TestLookupGraph(t *testing.T) {
	foo := &model.Graph{Name: "foo"}
	bar := &model.Graph{Name: "bar"}
	c := &server{
		loadedGraphs: map[string]*serveGraph{
			"foo": {Graph: foo},
			"bar": {Graph: bar},
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
		t.Run(test.key, func(t *testing.T) {
			g, err := c.lookupGraph(test.key)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.lookupGraph(%q) = %v, want %v", test.key, got, want)
			}
			if g == nil {
				return
			}
			if got, want := g.Graph, test.g; got != want {
				t.Errorf("c.lookupGraph(%q) = %v, want %v", test.key, got, want)
			}
		})
	}
}

func TestLookupNode(t *testing.T) {
	bar := &model.Node{Name: "bar"}
	g := &serveGraph{
		Graph: &model.Graph{
			Name:  "foo",
			Nodes: map[string]*model.Node{"bar": bar},
		},
	}
	tests := []struct {
		nk   string
		n    *model.Node
		code codes.Code
	}{
		{"bar", bar, codes.OK},
		{"baz", nil, codes.NotFound},
	}
	for _, test := range tests {
		t.Run(test.nk, func(t *testing.T) {
			n, err := g.lookupNode(test.nk)
			if got, want := code(err), test.code; got != want {
				t.Errorf("g.lookupNode(%q) = code %v, want %v", test.nk, got, want)
			}
			if got, want := n, test.n; got != want {
				t.Errorf("g.lookupNode(%q) = node %v, want %v", test.nk, got, want)
			}
		})
	}
}

func TestLookupChannel(t *testing.T) {
	bar := &model.Channel{Name: "bar"}
	g := &serveGraph{
		Graph: &model.Graph{
			Name:     "foo",
			Channels: map[string]*model.Channel{"bar": bar},
		},
	}
	tests := []struct {
		ck   string
		ch   *model.Channel
		code codes.Code
	}{
		{"bar", bar, codes.OK},
		{"baz", nil, codes.NotFound},
	}
	for _, test := range tests {
		t.Run(test.ck, func(t *testing.T) {
			ch, err := g.lookupChannel(test.ck)
			if got, want := code(err), test.code; got != want {
				t.Errorf("c.lookupChannel(%q) = code %v, want %v", test.ck, got, want)
			}
			if got, want := ch, test.ch; got != want {
				t.Errorf("c.lookupChannel(%q) = node %v, want %v", test.ck, got, want)
			}
		})
	}
}
