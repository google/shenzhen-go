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

type graphController struct {
	graph  *model.Graph
	client pb.ShenzhenGoClient
}

func (c *graphController) Graph() *model.Graph {
	return c.graph
}

func (c *graphController) CreateNode(ctx context.Context, partType string) error {
	// TODO
	return nil
}

func (c *graphController) RegisterOutlets() {
	// TODO
}

func (c *graphController) Save(ctx context.Context) error {
	_, err := c.client.Save(ctx, &pb.SaveRequest{Graph: c.graph.FilePath})
	return err
}

func (c *graphController) SaveProperties(ctx context.Context) error {
	// TODO
	return nil
}
