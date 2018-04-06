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
	"github.com/google/shenzhen-go/dev/client/view"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

type controller struct {
	graph  *model.Graph
	client pb.ShenzhenGoClient
}

// New returns a new controller for a graph.
func New(g *model.Graph, c pb.ShenzhenGoClient) view.Controller {
	return &controller{
		client: c,
		graph:  g,
	}
}

func (c *controller) GraphController() view.GraphController {
	return &graphController{
		client: c.client,
		graph:  c.graph,
	}
}

func (c *controller) PartTypes() map[string]*model.PartType {
	return model.PartTypes
}
