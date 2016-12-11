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

package main

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"text/template"
)

// There isn't a "parser.ParseArbitraryBlob" convenience function so we make do with
// ParseFile. This requires a "go file" complete with "package" and statements in a
// function body.
const wrapperTmplSrc = `package {{.RandIdent}}
func {{.RandIdent}}() {
    {{.Content}}
}`

var wrapperTmpl = template.Must(template.New("wrapper").Parse(wrapperTmplSrc))

type findFunc struct {
	funcName string
	subvis   ast.Visitor
}

func (v *findFunc) Visit(node ast.Node) ast.Visitor {
	f, ok := node.(*ast.FuncDecl)
	if !ok {
		return v
	}
	if f.Name.Name != v.funcName {
		return nil
	}
	return v.subvis
}

type chanIdents struct {
	srcs, dsts []string
}

func (v *chanIdents) Visit(node ast.Node) ast.Visitor {
	switch s := node.(type) {
	case *ast.SendStmt:
		id, ok := s.Chan.(*ast.Ident)
		if !ok {
			return v
		}
		v.dsts = append(v.dsts, id.Name)

	case *ast.UnaryExpr:
		if s.Op != token.ARROW {
			return nil
		}
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return nil
		}
		v.srcs = append(v.srcs, id.Name)

	case *ast.RangeStmt:
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return v
		}
		v.srcs = append(v.srcs, id.Name)

	}
	return v
}

func extractChannelIdents(src string) (srcs, dsts []string, err error) {
	fset := token.NewFileSet()
	rb := make([]byte, 8)
	if _, err = rand.Read(rb); err != nil {
		return nil, nil, err
	}
	rn := fmt.Sprintf("zzz%xzzz", rb)
	pr, pw := io.Pipe()
	go func() {
		wrapperTmpl.Execute(pw, struct{ RandIdent, Content string }{
			RandIdent: rn,
			Content:   src,
		})
		pw.Close()
	}()
	f, err := parser.ParseFile(fset, rn+".go", pr, 0)
	if err != nil {
		return nil, nil, err
	}
	ci := &chanIdents{}
	ast.Walk(&findFunc{funcName: rn, subvis: ci}, f)
	return ci.srcs, ci.dsts, nil
}
