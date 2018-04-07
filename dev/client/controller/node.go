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

	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

type nodeController struct {
	client pb.ShenzhenGoClient
	graph  *model.Graph
	node   *model.Node
}

func (c *nodeController) Node() *model.Node        { return c.node }
func (c *nodeController) Name() string             { return c.node.Name }
func (c *nodeController) Position() (x, y float64) { return c.node.X, c.node.Y }

func (c *nodeController) Delete(ctx context.Context) error {
	return nil // TODO
}

func (c *nodeController) Save(ctx context.Context) error {
	return nil // TODO
}
