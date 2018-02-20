// Copyright 2016 Google Inc.
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

package view

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/google/shenzhen-go/dev/model"
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
var (
	channelEditorTemplate = template.Must(template.New("channelEditor").Parse(string(templateResources["templates/channel.html"])))

	identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
)

// Channel displays the channel editor for a particular channel.
func Channel(w http.ResponseWriter, g *model.Graph, e *model.Channel, new bool) error {
	return channelEditorTemplate.Execute(w, &struct {
		*model.Graph
		*model.Channel
		New bool
	}{g, e, new})
}
