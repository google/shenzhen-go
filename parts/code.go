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

package parts

import (
	"html/template"
	"net/http"

	"github.com/google/shenzhen-go/source"
)

// Code is a component containing arbitrary code.
type Code struct {
	Code string `json:"code"`

	// Computed from Code - which channels are read from and written to.
	chansRd, chansWr []string
}

// AssociateEditor adds a "part_view" template to the given template.
func (c *Code) AssociateEditor(tmpl *template.Template) error {
	_, err := tmpl.New("part_view").Parse(`<textarea name="Code" rows="25" cols="80">{{.Node.Impl}}</textarea>`)
	return err
}

// Channels returns the names of all channels used by this goroutine.
func (c *Code) Channels() (read, written []string) { return c.chansRd, c.chansWr }

// Clone returns a copy of this Code part.
func (c *Code) Clone() interface{} {
	return &Code{
		Code:    c.Code,
		chansRd: c.chansRd,
		chansWr: c.chansWr,
	}
}

// Impl returns the implementation of the goroutine.
func (c *Code) Impl() string { return c.Code }

// Update sets relevant fields based on the given Request.
func (c *Code) Update(r *http.Request) error {
	code := c.Code
	if r != nil {
		code = r.FormValue("Code")
	}
	s, d, err := source.ExtractChannelIdents(code)
	if err != nil {
		return err
	}
	c.Code = code
	c.chansRd, c.chansWr = s, d
	return nil
}

// TypeKey returns "Code".
func (*Code) TypeKey() string { return "Code" }
