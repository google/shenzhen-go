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

package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type chanIdents struct {
	matchIdent    func(*ast.Ident) bool
	read, written StringSet
}

func (v *chanIdents) Visit(node ast.Node) ast.Visitor {
	switch s := node.(type) {
	case *ast.SendStmt:
		id, ok := s.Chan.(*ast.Ident)
		if !ok {
			return v
		}
		if !v.matchIdent(id) {
			return v
		}
		v.written.Add(id.Name)

	case *ast.UnaryExpr:
		if s.Op != token.ARROW {
			return nil
		}
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if !v.matchIdent(id) {
			return nil
		}
		v.read.Add(id.Name)

	case *ast.RangeStmt:
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return v
		}
		if !v.matchIdent(id) {
			return v
		}
		v.read.Add(id.Name)

	case *ast.CallExpr:
		// close(ch) is interpreted as writing to ch.
		if len(s.Args) != 1 {
			return v
		}
		fi, ok := s.Fun.(*ast.Ident)
		if !ok {
			return v
		}
		if fi.Name != "close" {
			return v
		}
		id, ok := s.Args[0].(*ast.Ident)
		if !ok {
			return v
		}
		if !v.matchIdent(id) {
			return v
		}
		v.written.Add(id.Name)
	}
	return v
}

// ExtractChannels extracts channel names used. The channel definitions are compared
// against definitions in defs, to avoid false positives (shadowed, ranging over non-channels).
func ExtractChannels(src, funcname, defs string) (read, written StringSet, err error) {
	fset := token.NewFileSet()
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}
	f, pkg, err := parseSnippet(src, funcname, defs, fset, 0, info)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing snippet: %v", err)
	}
	vars := make(map[string]types.Object)
	scope := pkg.Scope()
	for _, n := range scope.Names() {
		o := scope.Lookup(n)
		if _, ok := o.(*types.Var); !ok {
			continue
		}
		vars[n] = o
	}

	ci := &chanIdents{
		matchIdent: func(i *ast.Ident) bool {
			// Is i a usage of the package variable (of the same name)?
			return vars[i.Name] != nil && vars[i.Name] == info.Uses[i]
		},
		read:    make(StringSet),
		written: make(StringSet),
	}
	ast.Walk(&findFunc{funcName: funcname, subvis: ci}, f)
	return ci.read, ci.written, nil
}
