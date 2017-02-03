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
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
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
	f, err := parseSnippet(src, funcname, defs, fset, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parsing snippet: %v", err)
	}

	// To be extra sure we're adjusting the correct identifiers,
	// use go/types to resolve them all.
	cfg := types.Config{
		Error:    func(err error) { log.Printf("Typecheck error: %s", err) },
		Importer: importer.Default(),
	}
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}
	// Ignoring errors here since there's almost certainly going to be some
	// (any channel declarations for used channels that are not "from").
	pkg, _ := cfg.Check(funcname, fset, []*ast.File{f}, info)
	scope := pkg.Scope()
	fo := scope.Lookup(from)

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
	// The first three lines should be:
	//   package $funcname
	//
	//   func $funcname() {
	// and the last 4+count \n(defs) lines should be:
	//   }
	//
	//   $defs
	//
	// (The var line has a trailing \n, so four \n-separated segments)
	out = out[3:]
	out = out[:len(out)-4-strings.Count(defs, "\n")]

	// Each line in the function will be \t-indented 1 extra level.
	for i := range out {
		out[i] = strings.TrimPrefix(out[i], "\t")
	}
	return strings.Join(out, "\n"), nil
}
