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
	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
	pb "github.com/google/shenzhen-go/proto/go"
)

type pinController struct {
	client pb.ShenzhenGoClient
	graph  *model.Graph
	node   *model.Node
	name   string
	def    *pin.Definition
}

func (c *pinController) Name() string {
	return c.name
}

func (c *pinController) Type() string {
	return c.def.Type
}

func (c *pinController) IsInput() bool {
	return c.def.Direction == pin.Input
}

func (c *pinController) NodeName() string {
	return c.node.Name
}
