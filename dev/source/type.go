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

import (
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"regexp"
	"sort"
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

// Params returns a sorted slice of parameter names.
func (p *Type) Params() []TypeParam {
	params := make([]TypeParam, 0, len(p.paramToIdents))
	for param := range p.paramToIdents {
		params = append(params, param)
	}
	sort.Slice(params, func(i, j int) bool {
		if params[i].Scope == params[j].Scope {
			return params[i].Ident < params[j].Ident
		}
		return params[i].Scope < params[j].Scope
	})
	return params
}

// Refine fills in type parameters according to the provided map.
// If a parameter is not in the input map, it is left unrefined.
// If no parameters are in the input map, it does nothing.
func (p *Type) Refine(in map[TypeParam]*Type) error {
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
				return nil
			}
			if err := id.refine(subt.expr); err != nil {
				return err
			}
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
		return nil
	}
	p.spec = unmangle(types.ExprString(p.expr))
	return nil
}

// Lithify refines all parameters with a single default.
func (p *Type) Lithify(def *Type) error {
	// Quite similar to Refine.
	if p.Plain() {
		return nil
	}
	for _, ids := range p.paramToIdents {
		for _, id := range ids {
			if id.ident == p.expr {
				// Same reasoning as Refine.
				*p = *def
				return nil
			}
			if err := id.refine(def.expr); err != nil {
				return err
			}
		}
	}
	// Wholesale adopt all parameters, since all of
	// p's previous parameters were refined.
	p.identToParam = def.identToParam
	p.paramToIdents = def.paramToIdents
	p.spec = unmangle(types.ExprString(p.expr))
	return nil
}

type chanwalker struct {
	node chan ast.Node
	nxt  chan bool
}

func newChanwalker() *chanwalker {
	return &chanwalker{
		node: make(chan ast.Node),
		nxt:  make(chan bool),
	}
}

func (c *chanwalker) f(n ast.Node) bool {
	c.node <- n
	return <-c.nxt
}

func (c *chanwalker) inspect(e ast.Expr) {
	ast.Inspect(e, c.f)
	close(c.node)
}

func (c *chanwalker) close() {
	// Close next so f returns false, and then soak up any remaining nodes.
	close(c.nxt)
	for range c.node {
	}
}

func (c *chanwalker) next(b bool) {
	c.nxt <- b
}

// Infer attempts to produce a map `M` such that `p.Refine(M) = q`, or something
// more refined than q.
func (p *Type) Infer(q *Type) (map[TypeParam]*Type, error) {
	// Basic approach: Walk p.expr and t.expr at the same time.
	// If a meaningful difference is resolvable as a type parameter refinement, then
	// add it to the map, otherwise raise an error.
	pwalk, qwalk := newChanwalker(), newChanwalker()
	go pwalk.inspect(p.expr)
	go qwalk.inspect(q.expr)
	defer pwalk.close()
	defer qwalk.close()

	M := make(map[TypeParam]*Type)
	for pn := range pwalk.node {
		qn, ok := <-qwalk.node
		if !ok {
			// p has more nodes than q
			return nil, errors.New("types have mismatching shapes")
		}

		// Are either of pn or qn type parameters?
		pident, _ := pn.(*ast.Ident)
		qident, _ := qn.(*ast.Ident)
		tp, ppara := p.identToParam[pident]
		_, qpara := q.identToParam[qident]
		// Note qpara is true only if qident is not nil, etc.

		switch {
		case !ppara && !qpara:
			// Neither is; compare nodes as normal.
			if err := equal(pn, qn); err != nil {
				return nil, err
			}
			// could be anything, walk children.
			pwalk.next(true)
			qwalk.next(true)

		case ppara:
			// Note: do ppara case before qpara, because Infer should infer return
			// something for tp even if it is just a different parameter.
			// pn is a parameter and could match but first check qn is typeish.
			if !isType(qn) {
				return nil, fmt.Errorf("parameter %s cannot match non-type node %T", tp.Ident, qn)
			}
			// Unlike the qpara case below, we must remember what type matched pn.
			// It's a type so it fits in ast.Expr.
			qs := q.subtype(qn.(ast.Expr))

			// Did a refinement for tp already get inferred?
			// e.g. we inferred a type for the first $T in struct {F $T; G $T},
			// and just encountered the second $T.
			es := M[tp]
			if es == nil {
				// No.
				M[tp] = qs
				pwalk.next(false)
				qwalk.next(false)
				continue
			}
			// Yes. Are es and qs compatible? Recursive Infer can tell us.
			inf, err := es.Infer(qs)
			if err != nil {
				// Not compatible.
				return nil, err
			}
			// If len(inf) > 0, then qs is at least as specific than es.
			// TODO(josh): This could be improved if instead:
			//   es.Refine(inf)
			// but this would mutate es which is a previous subtype of q,
			// and I'd like Infer to be non-mutating.
			// But maybe it'd be fine.
			if len(inf) > 0 {
				M[tp] = qs
			}
			// param and ??, don't walk.
			pwalk.next(false)
			qwalk.next(false)

		case qpara:
			// qn is a paramter and could match, but first check pn is typeish.
			if !isType(pn) {
				return nil, fmt.Errorf("parameter %s cannot match non-type node %T", tp.Ident, qn)
			}

			// param and ??, so don't walk
			pwalk.next(false)
			qwalk.next(false)
		}
	}
	if _, ok := <-qwalk.node; ok {
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

var errIdentExpected = errors.New("want type parameter ident in parent node")

// refine finds id.ident inside id.parent and replaces it with subst.
// It only does this for parent type nodes (nodes that refer to a subtype).
func (id modIdent) refine(subst ast.Expr) error {
	switch par := id.parent.(type) {
	case *ast.ArrayType:
		if par.Elt != id.ident {
			return errIdentExpected
		}
		par.Elt = subst

	case *ast.ChanType:
		if par.Value != id.ident {
			return errIdentExpected
		}
		par.Value = subst

	case *ast.Field:
		// Covers structs, interfaces, and funcs (all contain FieldList).
		if par.Type != id.ident {
			return errIdentExpected
		}
		par.Type = subst

	case *ast.MapType:
		switch id.ident {
		case par.Key:
			par.Key = subst
		case par.Value:
			par.Value = subst
		default:
			return errIdentExpected
		}

	case *ast.StarExpr:
		// A pointer type (hopefully not a dereference expression).
		if par.X != id.ident {
			return errIdentExpected
		}
		par.X = subst

	default:
		return fmt.Errorf("cannot substitute into parent node type %T", id.parent)
	}
	return nil
}

func isType(n ast.Node) bool {
	switch n.(type) {
	case
		*ast.Ident,         // foo
		*ast.ArrayType,     // []foo
		*ast.ChanType,      // chan foo
		*ast.FuncType,      // func(a foo, b bar) baz
		*ast.InterfaceType, // interface { a() foo; b(bar) }
		*ast.MapType,       // map[foo]bar
		*ast.SelectorExpr,  // package.Foo
		*ast.StarExpr,      // *foo
		*ast.StructType:    // struct {a foo; b bar}
		// It's probably a type.
		return true
	default:
		return false
	}
}

func equal(m, n ast.Node) error {
	switch x := m.(type) {
	case *ast.ArrayType:
		_, ok := n.(*ast.ArrayType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Len and Elt should be walked.
	case *ast.BasicLit:
		// Can occur as, say, the Len of an array type.
		y, ok := n.(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		if x.Kind != y.Kind {
			return fmt.Errorf("basic literal kind mismatch [%v vs %v]", x.Kind, y.Kind)
		}
		xv := constant.MakeFromLiteral(x.Value, x.Kind, 0)
		yv := constant.MakeFromLiteral(y.Value, y.Kind, 0)
		if !constant.Compare(xv, token.EQL, yv) {
			return fmt.Errorf("basic literal not equal [%v vs %v]", xv, yv)
		}
	case *ast.ChanType:
		y, ok := n.(*ast.ChanType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		if x.Dir != y.Dir {
			return fmt.Errorf("chan type dir mismatch [%v vs %v]", x.Dir, y.Dir)
		}
	case *ast.Ellipsis:
		// Can be either an array len or function parameter list.
		_, ok := n.(*ast.Ellipsis)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
	case *ast.Field:
		_, ok := n.(*ast.Field)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Names, Type, and Tag should all be walked.
	case *ast.FieldList:
		_, ok := n.(*ast.FieldList)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// List should be walked.
	case *ast.FuncType:
		_, ok := n.(*ast.FuncType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Params and Results should be walked.
	case *ast.Ident:
		y, ok := n.(*ast.Ident)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		if x.Name != y.Name {
			return fmt.Errorf("idents not identical [%q vs %q]", x.Name, y.Name)
		}
	case *ast.InterfaceType:
		_, ok := n.(*ast.InterfaceType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Methods should be walked.
	case *ast.MapType:
		_, ok := n.(*ast.MapType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Key and Value should be walked.
	case *ast.SelectorExpr:
		_, ok := n.(*ast.SelectorExpr)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// X and Sel should be walked.
	case *ast.StarExpr:
		_, ok := n.(*ast.StarExpr)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// X should be walked.
	case *ast.StructType:
		_, ok := n.(*ast.StructType)
		if !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Fields should be walked.
	}
	return nil
}
