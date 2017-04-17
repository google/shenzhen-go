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
	"encoding/json"
	"go/format"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/shenzhen-go/model/pin"
)

const (
	codePartEditTemplateSrc = `
<script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.2.6/ace.js" type="text/javascript" charset="utf-8"></script>
<script>
    var theme = 'ace/theme/chrome';
	var lang = 'ace/mode/golang';
	var imps = ace.edit('imports');
    var head = ace.edit('head');
	var body = ace.edit('body');
	var tail = ace.edit('tail');
    imps.setTheme(theme);
    imps.getSession().setMode(lang);
	imps.getSession().setUseSoftTabs(false);
    head.setTheme(theme);
    head.getSession().setMode(lang);
	head.getSession().setUseSoftTabs(false);
    body.setTheme(theme);
    body.getSession().setMode(lang);
	body.getSession().setUseSoftTabs(false);
    tail.setTheme(theme);
    tail.getSession().setMode(lang);
	tail.getSession().setUseSoftTabs(false);
	this.parent.onsubmit = function() {
		document.getElementById('himports').value = imports.getValue();
		document.getElementById('hhead').value = head.getValue();
		document.getElementById('hbody').value = body.getValue();
		docuemnt.getElementById('htail').value = tail.getValue();
	};
</script>
`
)

var (
	// CodePanels are subpanels for editing code-type parts.
	CodePanels = []struct {
		Name   string
		Editor template.HTML
	}{
		{
			Name: "Pins",
			Editor: `<script>
function removemepls(e) {
	e.parentNode.removeChild(e);
}
function addrowpls() {
	var tr = document.createElement('tr');
	
	var c1 = document.createElement('td');
	tr.appendChild(c1);
	var dir = document.createElement('select');
	dir.name = 'PinDirection';
	c1.appendChild(dir);
	var inop = document.createElement('option');
	inop.value = 'In';
	inop.innerText = 'Input';
	dir.appendChild(inop);
	var outop = document.createElement('option');
	outop.value = 'Out';
	outop.innerText = 'Output';
	dir.appendChild(outop);

	var c2 = document.createElement('td');
	tr.appendChild(c2);
	var name = document.createElement('input');
	name.type = 'text';
	name.name = 'PinName';
	name.required = true;
	name.pattern = '^[_a-zA-Z][_a-zA-Z0-9]*$';
	name.title = 'Must start with a letter or underscore, and only contain letters, digits, or underscores.';
	c2.appendChild(name);

	var c3 = document.createElement('td');
	tr.appendChild(c3);
	var typ = document.createElement('input');
	typ.type = 'text';
	typ.name = 'PinType';
	typ.required = true;
	c3.appendChild(typ);
	
	var c4 = document.createElement('td');
	tr.appendChild(c4);
	var rem = document.createElement('a');
	rem.href = 'javascript:void(0)';
	rem.onclick = function() { removemepls(tr); };
	rem.innerText = 'Remove pin';
	c4.appendChild(rem);

	var pins = document.getElementById('pins');
	pins.appendChild(tr);
}
</script>
<table>
<thead>
	<tr>
		<th class="pin-col-1">Direction</th>
		<th class="pin-col-2">Name</th>
		<th class="pin-col-3">Type</th>
		<th class="pin-col-4"><a href="javascript:void(0)" onclick="addrowpls()">Add pin</a></th>
	</tr>
</thead>
<tbody id="pins">
{{range $.Node.Part.Pins}}
	<tr>
	    <td>
			<select name="PinDirection">
				<option value="in" {{if eq .Direction "in"}}selected{{end}}>Input</option>
				<option value="out" {{if eq .Direction "out"}}selected{{end}}>Output</option>
			</select>
		</td>
		<td>
			<input type="text" name="PinName" required pattern="^[_a-zA-Z][_a-zA-Z0-9]*$" title="Must start with a letter or underscore, and only contain letters, digits, or underscores." value="{{.Name}}">
		</td>
		<td>
			<input type="text" name="PinType" required value="{{.Type}}">
		</td>
		<td>
			<a href="javascript:void(0)" onclick="removemepls(this.parentNode.parentNode)">Remove pin</a>
		</td>
	</tr>
{{end -}}
</tbody>
</table>`,
		},
		{
			Name:   "Imports",
			Editor: `<pre class="codeedit" id="imports">{.Node.FlatImports}</pre>`,
		},
		{
			Name:   "Head",
			Editor: `<pre class="codeedit" id="head">{.Node.ImplHead}</pre>`,
		},
		{
			Name:   "Body",
			Editor: `<pre class="codeedit" id="body">{.Node.ImplBody}</pre>`,
		},
		{
			Name:   "Tail",
			Editor: `<pre class="codeedit" id="tail">{.Node.ImplTail}</pre>`,
		},
		{
			Name: "Help",
			Editor: `<p>
	A Code part runs (executes) any Go code that you write.
	</p><p>
	It consists of 3 parts: a Head, a Body, and a Tail. 
	The Head runs first, and only runs once, no matter what number Multiplicity is set to. 
	The Body runs next. The number of concurrent copies of the Body that run is set by Multiplicity.
	Finally, when all copies of the Body return, the Tail runs. 
	</p><p>
	The Head and Tail are useful for operations that should only be done once. For example, any 
	output channels written to in the Body can be correctly closed (if desired) in the Tail.
	</p><p>
	Each instance of the Body can use the int parameters <code>instanceNumber</code> and <code>multiplicity</code>
	to distinguish which instance is running and how many are running, if necessary. 
	<code>0 <= instanceNumber < multiplicity</code>
	</p><p>
	Any channels referred to will automatically be detected and shown in the graph, and
	when channels are renamed, these will be safely updated in the Code where they are
	referred to.
	</p><p>
	The <code>return</code> statement is allowed but optional in Code. There are no values that
	need to be returned.
	Using <code>return</code> in the Head will prevent the Body or Tail from executing, but 
	using <code>return</code> in the Body won't affect whether the Tail is executed.
	</p>
	`,
		},
	}
)

// Code is a component containing arbitrary code.
type Code struct {
	imports          []string
	head, body, tail string
	pins             []pin.Definition
}

// NewCode just makes a new *Code.
func NewCode(imports []string, head, body, tail string, pins []pin.Definition) *Code {
	return &Code{
		imports: imports,
		head:    head,
		body:    body,
		tail:    tail,
		pins:    pins,
	}
}

type part interface {
	Imports() []string
	Impl() (head, body, tail string)
	Pins() []pin.Definition
}

// NewCodeFromAny creates a new Code, copying the implementation details
// out of an existing part.
func NewCodeFromAny(p part) *Code {
	h, b, t := p.Impl()
	return &Code{
		imports: p.Imports(),
		head:    h,
		body:    b,
		tail:    t,
		pins:    p.Pins(),
	}
}

type jsonCode struct {
	Imports []string         `json:"imports"`
	Head    []string         `json:"head"`
	Body    []string         `json:"body"`
	Tail    []string         `json:"tail"`
	Pins    []pin.Definition `json:"pins"`
}

// MarshalJSON encodes the Code component as JSON.
func (c *Code) MarshalJSON() ([]byte, error) {
	k := &jsonCode{
		Imports: c.imports,
		Head:    strings.Split(c.head, "\n"),
		Body:    strings.Split(c.body, "\n"),
		Tail:    strings.Split(c.tail, "\n"),
		Pins:    c.pins,
	}
	stripCR(k.Imports)
	stripCR(k.Head)
	stripCR(k.Body)
	stripCR(k.Tail)
	return json.Marshal(k)
}

// UnmarshalJSON decodes the Code component from JSON.
func (c *Code) UnmarshalJSON(j []byte) error {
	var mp jsonCode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	h := strings.Join(mp.Head, "\n")
	b := strings.Join(mp.Body, "\n")
	t := strings.Join(mp.Tail, "\n")
	if err := c.refresh(h, b, t); err != nil {
		// TODO: revisit all this
		log.Printf("Couldn't format or determine channels used: %v", err)
	}
	c.imports = mp.Imports
	c.pins = mp.Pins
	return nil
}

// Pins returns pins. These are 100% user-defined.
func (c *Code) Pins() []pin.Definition { return c.pins }

// Clone returns a copy of this Code part.
func (c *Code) Clone() interface{} {
	c2 := &Code{
		imports: c.imports,
		head:    c.head,
		body:    c.body,
		tail:    c.tail,
		pins:    append([]pin.Definition{}, c.pins...),
	}
	return c2
}

// Impl returns the implementation of the goroutine.
func (c *Code) Impl() (Head, Body, Tail string) {
	return c.head, c.body, c.tail
}

// Imports returns a nil slice.
func (c *Code) Imports() []string { return c.imports }

// TypeKey returns "Code".
func (*Code) TypeKey() string { return "Code" }

func (c *Code) refresh(h, b, t string) error {
	// At least save what the user entered.
	c.head, c.body, c.tail = h, b, t

	// Try to format it.
	hf, err := format.Source([]byte(h))
	if err != nil {
		return err
	}
	bf, err := format.Source([]byte(b))
	if err != nil {
		return err
	}
	tf, err := format.Source([]byte(t))
	if err != nil {
		return err
	}

	c.head, c.body, c.tail = string(hf), string(bf), string(tf)
	return nil
}

// Update sets relevant fields based on the given Request.
func (c *Code) Update(r *http.Request) error {
	h, b, t := c.head, c.body, c.tail
	if r != nil {
		h, b, t = r.FormValue("Head"), r.FormValue("Body"), r.FormValue("Tail")
	}
	pd, pn, pt := r.Form["PinDirection"], r.Form["PinName"], r.Form["PinType"]
	c.pins = make([]pin.Definition, 0, len(pd))
	for i, d := range pd {
		c.pins = append(c.pins, pin.Definition{
			Name:      pn[i],
			Direction: pin.Direction(d),
			Type:      pt[i],
		})
	}
	return c.refresh(h, b, t)
}
