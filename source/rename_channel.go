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

package source

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
)

type renameIdent struct {
	matchFrom func(*ast.Ident) bool
	to        string
}

func (r *renameIdent) Visit(node ast.Node) ast.Visitor {
	i, ok := node.(*ast.Ident)
	if !ok {
		return r
	}
	if r.matchFrom(i) {
		i.Name = r.to
	}
	return r
}

// RenameChannel renames a package-level channel variable in a snippet of code.
func RenameChannel(src, funcname, from, to string) (string, error) {
	fset := token.NewFileSet()
	defs := fmt.Sprintf("var %s chan interface{}", from)
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}
	f, pkg, err := parseSnippet(src, funcname, defs, fset, parser.ParseComments, info)
	if err != nil {
		return "", fmt.Errorf("parsing snippet: %v", err)
	}
	fo := pkg.Scope().Lookup(from)

	rn := &renameIdent{
		matchFrom: func(i *ast.Ident) bool {
			return info.Uses[i] == fo
		},
		to: to,
	}
	ast.Walk(rn, f)
	buf := new(bytes.Buffer)

	if err := format.Node(buf, fset, f); err != nil {
		return "", fmt.Errorf("formatting output: %v", err)
	}

	out := strings.Split(buf.String(), "\n")
	// The first 4+N lines should be:
	//   package $funcname
	//
	//   $defs
	//
	//   func $funcname() {
	// and the last 2 lines should be:
	//   }
	//
	// (The } line has a trailing \n, so 2 \n-separated segments total)
	out = out[5+strings.Count(defs, "\n"):]
	out = out[:len(out)-2]

	// Each line in the function will be \t-indented 1 extra level.
	for i := range out {
		out[i] = strings.TrimPrefix(out[i], "\t")
	}
	return strings.Join(out, "\n"), nil
}
