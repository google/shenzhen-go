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
	srcs, dsts map[string]bool
}

func (v *chanIdents) Visit(node ast.Node) ast.Visitor {
	switch s := node.(type) {
	case *ast.SendStmt:
		id, ok := s.Chan.(*ast.Ident)
		if !ok {
			return v
		}
		v.dsts[id.Name] = true

	case *ast.UnaryExpr:
		if s.Op != token.ARROW {
			return nil
		}
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return nil
		}
		v.srcs[id.Name] = true

	case *ast.RangeStmt:
		id, ok := s.X.(*ast.Ident)
		if !ok {
			return v
		}
		v.srcs[id.Name] = true

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
		v.dsts[id.Name] = true

	}
	return v
}

func mapToSlice(m map[string]bool) (s []string) {
	s = make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return
}

// ExtractChannelIdents extracts identifier names which could be involved in
// channel reads (srcs) or writes (dsts). dsts only contains channel identifiers
// written to, but srcs can contain false positives.
func ExtractChannelIdents(src string) (srcs, dsts []string, err error) {
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
	ci := &chanIdents{srcs: make(map[string]bool), dsts: make(map[string]bool)}
	ast.Walk(&findFunc{funcName: rn, subvis: ci}, f)
	return mapToSlice(ci.srcs), mapToSlice(ci.dsts), nil
}
