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
)

type chanIdents struct {
	srcs, dsts StringSet
}

func (v *chanIdents) Visit(node ast.Node) ast.Visitor {
	switch s := node.(type) {
	case *ast.SendStmt:
		id, ok := s.Chan.(*ast.Ident)
		if !ok {
			return v
		}
		v.dsts.Add(id.Name)

	case *ast.UnaryExpr:
		if s.Op != token.ARROW {
			return nil
		}
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return nil
		}
		v.srcs.Add(id.Name)

	case *ast.RangeStmt:
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return v
		}
		v.srcs.Add(id.Name)

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
		v.dsts.Add(id.Name)
	}
	return v
}

// ExtractChannelIdents extracts identifier names which could be involved in
// channel reads (srcs) or writes (dsts). dsts only contains channel identifiers
// written to, but srcs can contain false positives.
func ExtractChannelIdents(src, funcname string) (srcs, dsts StringSet, err error) {
	fset := token.NewFileSet()
	f, err := parseSnippet(src, funcname, "", fset, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing snippet: %v", err)
	}
	ci := &chanIdents{srcs: make(map[string]struct{}), dsts: make(map[string]struct{})}
	ast.Walk(&findFunc{funcName: funcname, subvis: ci}, f)
	return ci.srcs, ci.dsts, nil
}
