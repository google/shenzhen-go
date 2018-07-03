// Copyright 2018 Google Inc.
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

// This is a bit crazy.
// Finding breaking examples is left as an exercise to the reader.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"regexp"
	"strings"
)

const (
	paramPrefix        = "$"
	mangledParamPrefix = "_SzGo_Mangled_Type_Param_" // Should be unlikely enough.
)

var mangledParamRE = regexp.MustCompile(regexp.QuoteMeta(mangledParamPrefix) + `\w+`)

// TypeParam identifies a type parameter in a pattern.
type TypeParam struct {
	Scope, Ident string
}

// Type represents a genericised type. Examples: $T, []$T, chan $T,
// map[$K]$V, struct { F $T }, and so on. It only understands types, it
// does not support genericised type _declarations_.
// Each type parameter may belong to a different scope after refinement,
// but each Type is constructed initially within a single scope.
type Type struct {
	spec          string
	expr          ast.Expr
	paramToIdents map[TypeParam][]modIdent
	identToParam  map[*ast.Ident]TypeParam
}

// These assume that paramPrefix is invalid in a regular Go type, which
// is false in general (struct tags).

func mangle(p string) string {
	return strings.Replace(p, paramPrefix, mangledParamPrefix, -1)
}

func unmangle(t string) string {
	return strings.Replace(t, mangledParamPrefix, paramPrefix, -1)
}

// NewType parses a generic type string into a Type.
// All parameters are assumed to belong to the one given scope.
// If t is not parametrised, scope is ignored.
func NewType(scope, t string) (*Type, error) {
	expr, err := parser.ParseExpr(mangle(t))
	if err != nil {
		return nil, err
	}
	if !isType(expr) {
		return nil, fmt.Errorf("parsed %q to non-type %T", t, expr)
	}
	identToParam := make(map[*ast.Ident]TypeParam)
	paramToIdents := make(map[TypeParam][]modIdent)
	pt := parentTracker{
		parent: nil,
		f: func(par, n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			if !mangledParamRE.MatchString(id.Name) {
				return true
			}
			tp := TypeParam{
				Scope: scope,
				Ident: unmangle(id.Name),
			}
			identToParam[id] = tp
			paramToIdents[tp] = append(paramToIdents[tp], modIdent{
				parent: par,
				ident:  id,
			})
			return false
		},
	}
	ast.Walk(pt, expr)
	return &Type{
		spec:          t,
		paramToIdents: paramToIdents,
		identToParam:  identToParam,
		expr:          expr,
	}, nil
}

// subtype returns a Type for the subexpression e.
func (p *Type) subtype(e ast.Expr) *Type {
	// Because p is already a type, its maps are already constructed.
	// We can inspect p to find idents that p knows are type params.
	// This is necessary for preserving scope.
	identToParam := make(map[*ast.Ident]TypeParam)
	paramToIdents := make(map[TypeParam][]modIdent)
	pt := parentTracker{
		parent: nil,
		f: func(par, n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			tp, ok := p.identToParam[id]
			if !ok {
				return false
			}
			identToParam[id] = tp
			paramToIdents[tp] = append(paramToIdents[tp], modIdent{
				parent: par,
				ident:  id,
			})
			return false
		},
	}
	ast.Walk(pt, e)
	return &Type{
		spec:          unmangle(types.ExprString(e)),
		paramToIdents: paramToIdents,
		identToParam:  identToParam,
		expr:          e,
	}
}

// Plain is true if the type has no parameters (is not generic).
func (p *Type) Plain() bool { return len(p.paramToIdents) == 0 }

// Params returns a slice of parameter names.
func (p *Type) Params() []TypeParam {
	params := make([]TypeParam, 0, len(p.paramToIdents))
	for param := range p.paramToIdents {
		params = append(params, param)
	}
	return params
}

// Refine fills in type parameters according to the provided map.
// If a parameter is not in the input map, it is left unrefined.
// If no parameters are in the input map, it does nothing.
func (p *Type) Refine(in map[TypeParam]*Type) {
	changed := false
	for tp, subt := range in {
		ids := p.paramToIdents[tp]
		if ids == nil {
			continue
		}
		delete(p.paramToIdents, tp)
		changed = true
		for _, id := range ids {
			if id.ident == p.expr {
				// Substitute the whole thing right now;
				// the whole of p is nothing but one type parameter.
				*p = *subt
				return
			}
			id.refine(subt.expr)
			delete(p.identToParam, id.ident)
			// And adopt subt's params.
			for sid, stp := range subt.identToParam {
				p.identToParam[sid] = stp
				if sid == subt.expr {
					// subt is just a parameter, but now its ident has a parent: whatever
					// id's parent was.
					p.paramToIdents[stp] = append(p.paramToIdents[stp], modIdent{
						parent: id.parent,
						ident:  sid,
					})
					break
				}
				// All of subt param should have parents inside subt.expr.
				p.paramToIdents[stp] = append(p.paramToIdents[stp], subt.paramToIdents[stp]...)
			}
		}
	}
	if !changed {
		return
	}
	p.spec = unmangle(types.ExprString(p.expr))
}

// Lithify refines all parameters with a single default.
func (p *Type) Lithify(def *Type) {
	// Quite similar to Refine.
	if p.Plain() {
		return
	}
	for _, ids := range p.paramToIdents {
		for _, id := range ids {
			if id.ident == p.expr {
				// Same reasoning as Refine.
				*p = *def
				return
			}
			id.refine(def.expr)
		}
	}
	// Wholesale adopt all parameters, since all of
	// p's previous parameters were refined.
	p.identToParam = def.identToParam
	p.paramToIdents = def.paramToIdents
	p.spec = unmangle(types.ExprString(p.expr))
}

// Infer attempts to produce a map `M` such that `p.Refine(M) = q`.
func (p *Type) Infer(q *Type) (map[TypeParam]*Type, error) {
	// Basic approach: Walk p.expr and t.expr at the same time.
	// If a meaningful difference is resolvable as a type parameter refinement, then
	// add it to the map, otherwise raise an error.
	pnode, qnode := make(chan ast.Node, 1), make(chan ast.Node, 1)
	pnext, qnext := make(chan bool), make(chan bool)
	go func() {
		ast.Inspect(p.expr, func(n ast.Node) bool {
			pnode <- n
			return <-pnext
		})
		close(pnode)
	}()
	go func() {
		ast.Inspect(q.expr, func(n ast.Node) bool {
			qnode <- n
			return <-qnext
		})
		close(qnode)
	}()
	defer func() {
		// On close the channels will read false forever, which wraps up ast.Inspect ASAP.
		close(pnext)
		close(qnext)
	}()

	M := make(map[TypeParam]*Type)
	for pn := range pnode {
		qn, ok := <-qnode
		if !ok {
			// p has more nodes than q
			return nil, errors.New("types have mismatching shapes")
		}

		// Is pn an ident?
		pident, ok := pn.(*ast.Ident)
		if !ok {
			// Basic comparison then.
			if err := meaningfullyEqual(pn, qn); err != nil {
				return nil, err
			}
		}
		// Is pn a type parameter of p?
		tp, ok := p.identToParam[pident]
		if !ok {
			// pn is a plain identifier, so qn should match name exactly.
			qident, ok := qn.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("cannot match plain ident with %T", qn)
			}
			if pident.Name != qident.Name {
				return nil, fmt.Errorf("mismatching idents [%q != %q]", pident.Name, qident.Name)
			}
		}

		// Is qn type-ish?
		if !isType(qn) {
			return nil, fmt.Errorf("parameter %s cannot match non-type node %T", tp.Ident, qn)
		}

		// Great! The whole qn can refine p in parameter tp.
		// It's a "type" per the above, so it fits in ast.Expr.
		// TODO: check for existing value in M[tp], if so, must match qn.
		M[tp] = q.subtype(qn.(ast.Expr))

		pnext <- false
		qnext <- false
	}

	if _, ok := <-qnode; ok {
		// q has more nodes than p
		return nil, errors.New("types have mismatching shapes")
	}

	return M, nil
}

func (p *Type) String() string {
	if p == nil {
		return "<unspecified>"
	}
	return p.spec
}

// parentTracker is an ast.Visitor that tracks the parent node of
// the node being visited.
type parentTracker struct {
	parent ast.Node
	f      func(parent, visit ast.Node) bool
}

func (t parentTracker) Visit(n ast.Node) ast.Visitor {
	if !t.f(t.parent, n) {
		return nil
	}
	return parentTracker{parent: n, f: t.f}
}

// modIdent holds enough information for substituting an ident for something else
// in an AST.
type modIdent struct {
	parent ast.Node
	ident  *ast.Ident
}

// refine finds id.ident inside id.parent and replaces it with subst.
// It only does this for parent type nodes (nodes that refer to a subtype).
func (id modIdent) refine(subst ast.Expr) {
	// TODO: check id.ident == par.Elt/Value/Type/Key etc
	switch par := id.parent.(type) {
	case *ast.ArrayType:
		par.Elt = subst
	case *ast.ChanType:
		par.Value = subst
	case *ast.Field:
		// Covers structs, interfaces, and funcs (all contain FieldList).
		par.Type = subst
	case *ast.MapType:
		if id.ident == par.Key {
			par.Key = subst
		}
		if id.ident == par.Value {
			par.Value = subst
		}
		// TODO: error on other types.
	}
}

func isType(n ast.Node) bool {
	switch n.(type) {
	case *ast.Ident, *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.StructType, *ast.InterfaceType:
		// It's probably a type.
		return true
	default:
		return false
	}
}

func meaningfullyEqual(m, n ast.Node) error {
	switch m.(type) {
	case *ast.ArrayType:
		_, ok := n.(*ast.ArrayType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// TODO: continue here
	}
	return nil
}
