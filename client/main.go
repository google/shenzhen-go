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

// The client script is for interacting with a graph in an SVG (via DOM).
// The browser gets it from the server (served "statically") and it makes
// gRPC-Web API calls into the server.
package main

import (
	"net/url"
	"strings"
	"syscall/js"

	"github.com/google/shenzhen-go/client/controller"
	"github.com/google/shenzhen-go/client/view"
	"github.com/google/shenzhen-go/dom"
	"github.com/google/shenzhen-go/model"
	_ "github.com/google/shenzhen-go/parts"
	pb "github.com/google/shenzhen-go/proto/js"
)

func main() {
	doc := dom.CurrentDocument()
	apiURL, err := url.Parse(doc.Get("baseURI").String())
	if err != nil {
		panic(err)
	}
	apiURL.Path = ""
	client := pb.NewShenzhenGoClient(apiURL.String())
	initial := js.Global().Get("graphJSON").String()
	graphPath := js.Global().Get("graphPath").String()
	g, err := model.LoadJSON(strings.NewReader(initial), graphPath, "")
	if err != nil {
		panic(err)
	}
	gc := controller.NewGraphController(doc, g, client)
	view.Setup(doc, gc)
}
