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
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"text/template"
)

// There isn't a "parser.ParseArbitraryBlob" convenience function so we make do with
// ParseFile. This requires a "go file" complete with "package" and statements in a
// function body.
const wrapperTmplSrc = `package {{.FuncName}}

{{.Defs}}

func {{.FuncName}}() { 
	{{.Content}}
}
`

var wrapperTmpl = template.Must(template.New("wrapper").Parse(wrapperTmplSrc))

func parseSnippet(src, funcname, defs string, fset *token.FileSet, mode parser.Mode, info *types.Info) (*ast.File, *types.Package, error) {
	pr, pw := io.Pipe()
	go func() {
		wrapperTmpl.Execute(pw, struct{ FuncName, Content, Defs string }{
			FuncName: funcname,
			Content:  src,
			Defs:     defs,
		})
		pw.Close()
	}()
	f, err := parser.ParseFile(fset, funcname+".go", pr, mode)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing file: %v", err)
	}
	// To be extra sure we're looking at the correct identifiers,
	// use go/types to resolve them all.
	cfg := types.Config{
		Error:    func(err error) {},
		Importer: importer.Default(),
	}
	// Ignoring errors here since there's almost certainly going to be some
	pkg, _ := cfg.Check(funcname, fset, []*ast.File{f}, info)
	return f, pkg, nil
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
