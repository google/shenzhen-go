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
	"strings"
)

type renameIdent struct {
	from, to string
}

func (r *renameIdent) Visit(node ast.Node) ast.Visitor {
	// Ignore "something.from"
	if _, ok := node.(*ast.SelectorExpr); ok {
		return nil
	}
	i, ok := node.(*ast.Ident)
	if !ok {
		return r
	}
	if i.Name == r.from {
		i.Name = r.to
	}
	return nil
}

// RenameIdent renames an identifier in a snippet of code.
func RenameIdent(src, funcname, from, to string) (string, error) {
	fset := token.NewFileSet()
	f, err := parseSnippet(src, funcname, fset, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parsing snippet: %v", err)
	}
	ff := &findFunc{funcName: funcname, subvis: &renameIdent{from: from, to: to}}
	ast.Walk(ff, f)
	buf := new(bytes.Buffer)

	if err := format.Node(buf, fset, f); err != nil {
		return "", fmt.Errorf("formatting output: %v", err)
	}

	out := strings.Split(buf.String(), "\n")
	// The first three lines should be
	//   package $funcname
	//
	//   func $funcname() {
	// and the last two lines should be
	//   }
	//
	// (The final } has a trailing \n)
	out = out[3:]
	out = out[:len(out)-2]

	// Each line in the function will be \t-indented 1 extra level.
	for i := range out {
		out[i] = strings.TrimPrefix(out[i], "\t")
	}
	return strings.Join(out, "\n"), nil
}
