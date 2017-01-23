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

// Package source helps deal with Go source code.
package source

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"text/template"
)

// There isn't a "parser.ParseArbitraryBlob" convenience function so we make do with
// ParseFile. This requires a "go file" complete with "package" and statements in a
// function body.
//
// This puts the package and function declaration on the first line. This should
// preserve line numbers for any errors (a trick learned from
// golang.org/x/tools/imports/imports.go).
const wrapperTmplSrc = "package {{.FuncName}}; func {{.FuncName}}() { {{.Content}} \n}"

var wrapperTmpl = template.Must(template.New("wrapper").Parse(wrapperTmplSrc))

func parseSnippet(src, funcname string, fset *token.FileSet, mode parser.Mode) (*ast.File, error) {
	pr, pw := io.Pipe()
	go func() {
		wrapperTmpl.Execute(pw, struct{ FuncName, Content string }{
			FuncName: funcname,
			Content:  src,
		})
		pw.Close()
	}()
	return parser.ParseFile(fset, funcname+".go", pr, mode)
}

type findFunc struct {
	funcName string
	subvis   ast.Visitor
	node     *ast.FuncDecl
}

func (v *findFunc) Visit(node ast.Node) ast.Visitor {
	f, ok := node.(*ast.FuncDecl)
	if !ok {
		return v
	}
	if f.Name.Name != v.funcName {
		return nil
	}
	v.node = f
	return v.subvis
}
