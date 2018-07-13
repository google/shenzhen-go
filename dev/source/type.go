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
	"sort"
	"strings"
)

const (
	paramPrefix        = "$"
	mangledParamPrefix = "_SzGo_Mangled_Type_Param_" // Should be unlikely enough.
)

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

// mangle/unmangle assume that paramPrefix is invalid in a regular Go type, which
// is false in general (struct tags).
// Typically a snippet is round-tripped through both mangle and unmangle, which undoes
// any unintended damage to paramPrefix, but causes problems if mangledParamPrefix is
// somehow used in the original snippet.

func mangle(p string) string {
	return strings.Replace(p, paramPrefix, mangledParamPrefix, -1)
}

func unmangle(t string) string {
	return strings.Replace(t, mangledParamPrefix, paramPrefix, -1)
}

func mangleIdent(n string) string {
	return mangledParamPrefix + strings.TrimPrefix(n, paramPrefix)
}

func unmangleIdent(n string) string {
	return paramPrefix + strings.TrimPrefix(n, mangledParamPrefix)
}

// MustNewType is NewType but where all errors cause a panic.
func MustNewType(scope, t string) *Type {
	typ, err := NewType(scope, t)
	if err != nil {
		panic(err)
	}
	return typ
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
			if !strings.HasPrefix(id.Name, mangledParamPrefix) {
				return true
			}
			tp := TypeParam{
				Scope: scope,
				Ident: unmangleIdent(id.Name),
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

// clone returns a deep copy of p, unless it is nil or plain.
// clone is only needed for parametrised types to prevent
// parameter forgetfulness.
func (p *Type) clone() *Type {
	if p == nil || p.Plain() {
		return p
	}
	q := &Type{
		spec:          p.spec,
		paramToIdents: make(map[TypeParam][]modIdent),
		identToParam:  make(map[*ast.Ident]TypeParam),
		expr:          nil,
	}
	ast.Walk(cloneWalker{oldtype: p, newtype: q}, p.expr)
	return q
}

type cloneWalker struct {
	oldtype, newtype *Type
	oldpar, newpar   ast.Node
}

func (cw cloneWalker) Visit(m ast.Node) ast.Visitor {
	if m == nil {
		return nil
	}

	n := shallowCopy(m)

	// If m is actually a parameter, n is a parameter in the new type.
	if id, ok := m.(*ast.Ident); ok {
		if tp, ok := cw.oldtype.identToParam[id]; ok {
			nid := n.(*ast.Ident)
			cw.newtype.identToParam[nid] = tp
			cw.newtype.paramToIdents[tp] = append(cw.newtype.paramToIdents[tp], modIdent{
				parent: cw.newpar,
				ident:  nid,
			})
		}
	}

	switch op := cw.oldpar.(type) {
	case *ast.ArrayType:
		np := cw.newpar.(*ast.ArrayType)
		switch m {
		case op.Len:
			np.Len = n.(ast.Expr)
		case op.Elt:
			np.Elt = n.(ast.Expr)
		}

	case *ast.ChanType:
		np := cw.newpar.(*ast.ChanType)
		np.Value = n.(ast.Expr)

	case *ast.Field:
		np := cw.newpar.(*ast.Field)
		np.Type = n.(ast.Expr)

	case *ast.FieldList:
		np := cw.newpar.(*ast.FieldList)
		// Yuck.
		for i, f := range op.List {
			if m == f {
				np.List[i] = n.(*ast.Field)
			}
		}

	case *ast.FuncType:
		np := cw.newpar.(*ast.FuncType)
		switch m {
		case op.Params:
			np.Params = n.(*ast.FieldList)
		case op.Results:
			np.Results = n.(*ast.FieldList)
		}

	case *ast.InterfaceType:
		np := cw.newpar.(*ast.InterfaceType)
		np.Methods = n.(*ast.FieldList)

	case *ast.MapType:
		np := cw.newpar.(*ast.MapType)
		switch m {
		case op.Key:
			np.Key = n.(ast.Expr)
		case op.Value:
			np.Value = n.(ast.Expr)
		}

	case *ast.ParenExpr:
		np := cw.newpar.(*ast.ParenExpr)
		np.X = n.(ast.Expr)

	case *ast.SelectorExpr:
		np := cw.newpar.(*ast.SelectorExpr)
		switch m {
		case op.Sel:
			np.Sel = n.(*ast.Ident)
		case op.X:
			np.X = n.(ast.Expr)
		}

	case *ast.StarExpr:
		np := cw.newpar.(*ast.StarExpr)
		np.X = n.(ast.Expr)

	case *ast.StructType:
		np := cw.newpar.(*ast.StructType)
		np.Fields = n.(*ast.FieldList)

	default:
		if op != nil {
			panic(fmt.Sprintf("unsupported ast.Node type %T", n))
		}
		// Only the initial cw starts with nils.
		cw.newtype.expr = n.(ast.Expr)

	}
	if !isParentOfType(m) {
		return nil
	}
	return cloneWalker{
		oldtype: cw.oldtype,
		newtype: cw.newtype,
		oldpar:  m,
		newpar:  n,
	}
}

// subtype returns a Type for the subexpression e.
func (p *Type) subtype(e ast.Expr) *Type {
	// Take advantage of cloning technology.
	return (&Type{
		spec:          unmangle(types.ExprString(e)),
		paramToIdents: p.paramToIdents,
		identToParam:  p.identToParam,
		expr:          e,
	}).clone()
}

// Plain is true if the type has no parameters (is not generic).
func (p *Type) Plain() bool { return len(p.paramToIdents) == 0 }

// Params returns a sorted slice of parameter names.
func (p *Type) Params() []TypeParam {
	if p == nil {
		return nil
	}
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
// It returns true if the refinement had any effect.
// If a parameter is not in the input map, it is left unrefined.
// If no parameters are in the input map, it does nothing.
func (p *Type) Refine(in TypeInferenceMap) (bool, error) {
	if p == nil {
		return false, nil
	}

	q := make(TypeInferenceMap)
	for tp := range p.paramToIdents {
		if st := in[tp]; st != nil {
			q[tp] = st
		}
	}

	changed := len(q) > 0
	for len(q) > 0 {
		tp, st := q.any1()
		delete(q, tp)

		if err := p.refine1(tp, st.clone()); err != nil {
			return true, err
		}

		for tp := range st.paramToIdents {
			if st := in[tp]; st != nil {
				q[tp] = st
			}
		}
	}
	if changed {
		p.spec = unmangle(types.ExprString(p.expr))
	}
	return changed, nil
}

func (p *Type) refine1(tp TypeParam, subst *Type) error {
	ids := p.paramToIdents[tp]
	delete(p.paramToIdents, tp)
	for _, id := range ids {
		if id.ident == p.expr {
			// Substitute the whole thing right now;
			// the whole of p is nothing but one type parameter.
			*p = *subst
			return nil
		}
		if err := id.refine(subst.expr); err != nil {
			return err
		}
		delete(p.identToParam, id.ident)
		// And adopt subt's params.
		for sid, stp := range subst.identToParam {
			p.identToParam[sid] = stp
			if sid == subst.expr {
				// subt is just a parameter, but now its ident has a parent: whatever
				// id's parent was.
				p.paramToIdents[stp] = append(p.paramToIdents[stp], modIdent{
					parent: id.parent,
					ident:  sid,
				})
				break
			}
			// All of subst params should have parents inside subst.expr.
			p.paramToIdents[stp] = append(p.paramToIdents[stp], subst.paramToIdents[stp]...)
		}
	}
	return nil
}

func (p *Type) String() string {
	if p == nil {
		return "<unspecified>"
	}
	return p.spec
}

type chanwalker struct {
	node chan ast.Node
	nxt  chan bool
}

func newChanwalker() *chanwalker {
	return &chanwalker{
		node: make(chan ast.Node, 1),
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

// TypeInferenceMap (TypeParam -> *Type) holds inferences made about
// type parameters.s
type TypeInferenceMap map[TypeParam]*Type

// Note ensures type params from t are keys in the map.
func (m TypeInferenceMap) Note(t *Type) {
	for tp := range t.paramToIdents {
		if _, found := m[tp]; !found {
			m[tp] = nil
		}
	}
}

// Infer attempts to add inferences to the map `m` such that `p.Refine(m)` and `q.Refine(m)`
// are similar. It returns an error if this is impossible.
func (m TypeInferenceMap) Infer(p, q *Type) error {
	// Basic approach: Walk p.expr and q.expr at the same time.
	// If a meaningful difference is resolvable as a type parameter refinement, then
	// add it to the map, otherwise raise an error.
	pwalk, qwalk := newChanwalker(), newChanwalker()
	go pwalk.inspect(p.expr)
	go qwalk.inspect(q.expr)
	defer pwalk.close()
	defer qwalk.close()

	for {
		pn, pk := <-pwalk.node
		qn, qk := <-qwalk.node
		if pk != qk {
			return errors.New("types have mismatching shapes")
		}
		if !pk { // and !qk, by the above.
			return nil
		}

		w, err := m.inferAtNode(p, q, pn, qn)
		if err != nil {
			return err
		}
		pwalk.next(w)
		qwalk.next(w)
	}
}

func (m TypeInferenceMap) inferAtNode(p, q *Type, pn, qn ast.Node) (bool, error) {
	// Are either of pn or qn type parameters?
	pident, _ := pn.(*ast.Ident)
	qident, _ := qn.(*ast.Ident)
	tp, ppara := p.identToParam[pident]
	tq, qpara := q.identToParam[qident]
	// Note qpara is true only if qident is not nil, etc.

	switch {
	case ppara && qpara:
		// We get nowhere by inferring that the two parameters are equal, so
		// drop it.
		return false, nil

	case ppara:
		// pn is a parameter and could match but first check qn is typeish.
		if !isType(qn) {
			return false, fmt.Errorf("parameter %s cannot match non-type node %T", tp.Ident, qn)
		}
		// It's a type or expr, so it fits in ast.Expr.
		qs := q.subtype(qn.(ast.Expr))
		return false, m.learn(tp, qs)

	case qpara:
		// qn is a paramter and could match, but first check pn is typeish.
		if !isType(pn) {
			return false, fmt.Errorf("parameter %s cannot match non-type node %T", tp.Ident, qn)
		}

		ps := p.subtype(pn.(ast.Expr))
		return false, m.learn(tq, ps)

	default:
		// Neither is; compare nodes as normal, and walk all children.
		return true, isEqual(pn, qn)
	}
}

func (m TypeInferenceMap) learn(tp TypeParam, st *Type) error {
	// Quick check: is tp a parameter of st? That's a recursive type (disallowed).
	if _, para := st.paramToIdents[tp]; para {
		return fmt.Errorf("inferred type recursion [%s in %s]", tp.Ident, st)
	}

	// Did a refinement for tp already get inferred?
	// e.g. we inferred a type for the first $T in struct {F $T; G $T},
	// and just encountered the second $T.
	et := m[tp]
	if et == nil {
		// No.
		m[tp] = st
		return nil
	}
	// Yes. Are et and qt compatible? Recursive Infer can tell us, and
	// learn yet more inferences.
	return m.Infer(et, st)
}

func (m TypeInferenceMap) any1() (TypeParam, *Type) {
	for tp, st := range m {
		return tp, st
	}
	return TypeParam{}, nil
}

// ApplyDefault sets all keys associated with a nil type to a given
// default type.
func (m TypeInferenceMap) ApplyDefault(t *Type) {
	for p, v := range m {
		if v == nil {
			m[p] = t
		}
	}
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

	case *ast.SelectorExpr:
		// Only if subst is an ident.
		if par.Sel != id.ident {
			return errIdentExpected
		}
		si, ok := subst.(*ast.Ident)
		if !ok {
			return errors.New("must substitute an ident only in selector expressions")
		}
		par.Sel = si

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
	switch x := n.(type) {
	case
		*ast.Ident,         // foo
		*ast.ArrayType,     // []foo
		*ast.ChanType,      // chan foo
		*ast.FuncType,      // func(a foo, b bar) baz
		*ast.InterfaceType, // interface { a() foo; b(bar) }
		*ast.MapType,       // map[foo]bar
		*ast.StarExpr,      // *foo
		*ast.StructType:    // struct {a foo; b bar}
		// It's probably a type.
		return true

	case *ast.SelectorExpr: // X.Foo
		// X must be an identifier.
		_, ok := x.X.(*ast.Ident)
		return ok

	case *ast.ParenExpr: // (foo)
		// X must be a type.
		return isType(x.X)

	default:
		return false
	}
}

func isParentOfType(n ast.Node) bool {
	switch x := n.(type) {
	case
		*ast.ArrayType,
		*ast.ChanType,
		*ast.Field,
		*ast.FieldList,
		*ast.FuncType,
		*ast.InterfaceType,
		*ast.MapType,
		*ast.StarExpr,
		*ast.StructType:
		return true

	case *ast.SelectorExpr: // X.Foo
		// X must be an identifier.
		_, ok := x.X.(*ast.Ident)
		return ok

	case *ast.ParenExpr: // (foo)
		// X must be a type.
		return isType(x.X)

	default:
		return false
	}
}

func isEqual(m, n ast.Node) error {
	if (m == nil) != (n == nil) {
		return fmt.Errorf("mismatching nils [%#v vs %#v]", m, n)
	}
	switch x := m.(type) {
	case *ast.ArrayType:
		if _, ok := n.(*ast.ArrayType); !ok {
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
		if constant.Compare(xv, token.NEQ, yv) {
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
		if _, ok := n.(*ast.Ellipsis); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
	case *ast.Field:
		if _, ok := n.(*ast.Field); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Names, Type, and Tag should all be walked.
	case *ast.FieldList:
		if _, ok := n.(*ast.FieldList); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// List should be walked.
	case *ast.FuncType:
		if _, ok := n.(*ast.FuncType); !ok {
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
		if _, ok := n.(*ast.InterfaceType); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Methods should be walked.
	case *ast.MapType:
		if _, ok := n.(*ast.MapType); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Key and Value should be walked.
	case *ast.ParenExpr:
		// Maybe they parenthesised a type for emphasis.
		if _, ok := n.(*ast.ParenExpr); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// X should be walked.
	case *ast.SelectorExpr:
		if _, ok := n.(*ast.SelectorExpr); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// X and Sel should be walked.
	case *ast.StarExpr:
		if _, ok := n.(*ast.StarExpr); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// X should be walked.
	case *ast.StructType:
		if _, ok := n.(*ast.StructType); !ok {
			return fmt.Errorf("node type mismatch [%T vs %T]", m, n)
		}
		// Fields should be walked.
	}
	return nil
}

func shallowCopy(n ast.Node) ast.Node {
	if n == nil {
		return nil
	}
	switch x := n.(type) {
	case *ast.ArrayType:
		y := *x
		return &y
	case *ast.BasicLit:
		y := *x
		return &y
	case *ast.ChanType:
		y := *x
		return &y
	case *ast.Ellipsis:
		y := *x
		return &y
	case *ast.Field:
		y := *x
		return &y
	case *ast.FieldList:
		y := *x
		// Don't want to overwrite fields in x, so...
		y.List = make([]*ast.Field, len(x.List))
		return &y
	case *ast.FuncType:
		y := *x
		return &y
	case *ast.Ident:
		y := *x
		return &y
	case *ast.InterfaceType:
		y := *x
		return &y
	case *ast.MapType:
		y := *x
		return &y
	case *ast.ParenExpr:
		y := *x
		return &y
	case *ast.SelectorExpr:
		y := *x
		return &y
	case *ast.StarExpr:
		y := *x
		return &y
	case *ast.StructType:
		y := *x
		return &y
	default:
		panic(fmt.Sprintf("unsupported ast.Node type %T", n))
	}
}

/*
// Expand expands type parameters in a single Go file. All type parameters
// are assumed to be in scope for the file.
func Expand(filename, src string, types map[string]*Type) (string, error) {
	mangledSrc := mangle(src)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, mangledSrc, 0)
	if err != nil {
		return "", err
	}

	pt := parentTracker{parent: nil, f: func(parent, node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		if !strings.HasPrefix(ident.Name, mangledParamPrefix) {
			return false
		}
		typ := types[unmangleIdent(ident.Name)]
		if typ == nil {
			return false
		}

		// TODO(josh): finish
		return false
	}}
	ast.Walk(pt, f)

	return "", nil
}
*/
