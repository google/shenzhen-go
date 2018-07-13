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
	"fmt"
	"testing"

	"gopkg.in/d4l3k/messagediff.v1"
)

func TestNewType(t *testing.T) {
	tests := []struct {
		scope      string
		spec       string
		wantPlain  bool
		wantIdents int
		wantParams []TypeParam
	}{
		{
			scope:      "foo",
			spec:       "int",
			wantPlain:  true,
			wantIdents: 0,
			wantParams: []TypeParam{},
		},
		{
			scope:      "bar",
			spec:       "string",
			wantPlain:  true,
			wantIdents: 0,
			wantParams: []TypeParam{},
		},
		{
			scope:      "baz",
			spec:       "map[string]map[int]struct{F bar;G interface{}}",
			wantPlain:  true,
			wantIdents: 0,
			wantParams: []TypeParam{},
		},
		{
			scope:      "boop",
			spec:       "somepackage.Type",
			wantPlain:  true,
			wantIdents: 0,
			wantParams: []TypeParam{},
		},
		{
			scope:      "foo",
			spec:       "$T",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"foo", "$T"}},
		},
		{
			scope:      "bar",
			spec:       "$barnacle777",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"bar", "$barnacle777"}},
		},
		{
			scope:      "baz",
			spec:       "*$T",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"baz", "$T"}},
		},
		{
			scope:      "qux",
			spec:       "[]$T",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"qux", "$T"}},
		},
		{
			scope:      "axk",
			spec:       "map[$K]$V",
			wantPlain:  false,
			wantIdents: 2,
			wantParams: []TypeParam{{"axk", "$K"}, {"axk", "$V"}},
		},
		{
			scope:      "tuz",
			spec:       "map[$T]$T",
			wantPlain:  false,
			wantIdents: 2,
			wantParams: []TypeParam{{"tuz", "$T"}},
		},
		{
			scope:      "foo",
			spec:       "struct { F $T }",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"foo", "$T"}},
		},
		{ // Here's one I didn't imagine.
			scope:      "foo",
			spec:       "somepackage.$T",
			wantPlain:  false,
			wantIdents: 1,
			wantParams: []TypeParam{{"foo", "$T"}},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s/%s", test.scope, test.spec), func(t *testing.T) {
			got, err := NewType(test.scope, test.spec)
			if err != nil {
				t.Fatalf("NewType(%s, %s) = error %v", test.scope, test.spec, err)
			}
			if got, want := got.Plain(), test.wantPlain; got != want {
				t.Errorf("Plain() = %v, want %v", got, want)
			}
			if got, want := len(got.identToParam), test.wantIdents; got != want {
				t.Errorf("len(identToParam) = %v, want %v", got, want)
			}
			if diff, equal := messagediff.PrettyDiff(got.Params(), test.wantParams); !equal {
				t.Errorf("Params() diff\n%s", diff)
			}
		})
	}
}

func TestRefine(t *testing.T) {
	tests := []struct {
		base *Type
		in   TypeInferenceMap
		want string
	}{
		{
			base: MustNewType("foo", "struct{}"),
			in: TypeInferenceMap{
				{"foo", "$T"}: MustNewType("", "int"),
			},
			want: "struct{}",
		},
		{
			base: MustNewType("foo", "$T"),
			in: TypeInferenceMap{
				{"foo", "$T"}: MustNewType("", "int"),
			},
			want: "int",
		},
		{
			base: MustNewType("bar", "$U"),
			in: TypeInferenceMap{
				{"bar", "$U"}: MustNewType("", "string"),
			},
			want: "string",
		},
		{
			base: MustNewType("foo", "$T"),
			in: TypeInferenceMap{
				{"bar", "$U"}: MustNewType("", "string"),
			},
			want: "$T",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: TypeInferenceMap{
				{"foo", "$K"}: MustNewType("", "string"),
				{"foo", "$V"}: MustNewType("", "int"),
			},
			want: "map[string]int",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: TypeInferenceMap{
				{"foo", "$V"}: MustNewType("", "int"),
			},
			want: "map[$K]int",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: TypeInferenceMap{
				{"foo", "$K"}: MustNewType("", "int"),
			},
			want: "map[int]$V",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: TypeInferenceMap{
				{"foo", "$K"}: MustNewType("bar", "struct{ F $T }"),
				{"foo", "$V"}: MustNewType("baz", "[]*$U"),
			},
			want: "map[struct{F $T}][]*$U",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: TypeInferenceMap{
				{"foo", "$K"}: MustNewType("bar", "struct{ F $T }"),
				{"foo", "$V"}: MustNewType("baz", "[]*$U"),
				{"bar", "$T"}: MustNewType("", "int"),
				{"baz", "$U"}: MustNewType("", "string"),
			},
			want: "map[struct{F int}][]*string",
		},
	}

	for _, test := range tests {
		t.Run(test.base.String(), func(t *testing.T) {
			// TODO(josh): test the changed bool.
			if _, err := test.base.Refine(test.in); err != nil {
				t.Fatalf("base(%s).Refine(%v) = error %v", test.base, test.in, err)
			}
			if got, want := test.base.String(), test.want; got != want {
				t.Errorf("base = %s, want %s", got, want)
			}
		})
	}
}

func TestInfer(t *testing.T) {
	tests := []struct {
		p, q *Type
		want map[TypeParam]string
	}{
		{
			p:    MustNewType("foo", "int"),
			q:    MustNewType("", "int"),
			want: map[TypeParam]string{},
		},
		{
			p:    MustNewType("foo", "packaged.Type"),
			q:    MustNewType("", "packaged.Type"),
			want: map[TypeParam]string{},
		},
		{
			p: MustNewType("foo", "$T"),
			q: MustNewType("", "int"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "int",
			},
		},
		{
			p: MustNewType("foo", "*$T"),
			q: MustNewType("", "*int"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "int",
			},
		},
		{
			p: MustNewType("bar", "[]$T"),
			q: MustNewType("", "[]string"),
			want: map[TypeParam]string{
				{"bar", "$T"}: "string",
			},
		},
		{
			p: MustNewType("foo", "map[$K]$V"),
			q: MustNewType("", "map[interface{}]struct{}"),
			want: map[TypeParam]string{
				{"foo", "$K"}: "interface{}",
				{"foo", "$V"}: "struct{}",
			},
		},
		{
			p: MustNewType("foo", "packaged.$T"),
			q: MustNewType("", "packaged.Type"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "Type",
			},
		},
		{
			p: MustNewType("foo", "$T"),
			q: MustNewType("", "packaged.Type"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "packaged.Type",
			},
		},
		{
			p: MustNewType("foo", "struct{F $T; G $U}"),
			q: MustNewType("", "struct { F float64; G complex128 }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "float64",
				{"foo", "$U"}: "complex128",
			},
		},
		{
			p: MustNewType("foo", "map[$T]$T"),
			q: MustNewType("", "map[interface{}]interface{}"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "interface{}",
			},
		},
		{ // Parameter equality is not recorded.
			p:    MustNewType("foo", "$T"),
			q:    MustNewType("bar", "$U"),
			want: map[TypeParam]string{},
		},
		{
			p: MustNewType("foo", "map[$K]string"),
			q: MustNewType("bar", "map[int]$V"),
			want: map[TypeParam]string{
				{"foo", "$K"}: "int",
				{"bar", "$V"}: "string",
			},
		},
		{
			p: MustNewType("foo", "struct{F $T; G $T}"),
			q: MustNewType("bar", "struct { F map[$K]$V; G map[string]int }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[$K]$V",
				{"bar", "$K"}: "string",
				{"bar", "$V"}: "int",
			},
		},
		{
			p: MustNewType("foo", "struct{F $T; G $T}"),
			q: MustNewType("bar", "struct { F map[string]int; G map[$K]$V }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[string]int",
				{"bar", "$K"}: "string",
				{"bar", "$V"}: "int",
			},
		},
		{
			p: MustNewType("foo", "struct{F $T; G $T}"),
			q: MustNewType("bar", "struct { F map[$K]int; G map[string]$V }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[$K]int",
				{"bar", "$K"}: "string",
				{"bar", "$V"}: "int",
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s+%s", test.p, test.q), func(t *testing.T) {
			m := make(TypeInferenceMap)
			if err := m.Infer(test.p, test.q); err != nil {
				t.Fatalf("Infer(%s, %s) = error %v", test.p, test.q, err)
			}
			got := make(map[TypeParam]string)
			for param, typ := range m {
				got[param] = typ.String()
			}
			if diff, equal := messagediff.PrettyDiff(got, test.want); !equal {
				t.Errorf("inferred map diff:\n%s", diff)
			}
		})
	}
}

func TestInferErrors(t *testing.T) {
	tests := []struct {
		p, q *Type
	}{
		{ // Plain not equal
			p: MustNewType("foo", "int"),
			q: MustNewType("bar", "string"),
		},
		{ // Mismatching container vs ident
			p: MustNewType("foo", "[]$T"),
			q: MustNewType("bar", "complex128"),
		},
		{ // Mismatching container types
			p: MustNewType("foo", "[]$T"),
			q: MustNewType("bar", "map[$K]$V"),
		},
		{ // Mismatching array lengths
			p: MustNewType("foo", "[3]$T"),
			q: MustNewType("bar", "[4]$U"),
		},
		{ // Type recursion
			p: MustNewType("foo", "$T"),
			q: MustNewType("foo", "[]$T"),
		},
		{ // Type recursion II
			p: MustNewType("foo", "$T"),
			q: MustNewType("foo", "*$T"),
		},
		{ // Type recursion III
			p: MustNewType("foo", "map[$T]$V"),
			q: MustNewType("foo", "$T"),
		},
		{ // Type recursion IV
			p: MustNewType("foo", "$T"),
			q: MustNewType("foo", "interface{ F() struct{ M map[$K][]$T }}"),
		},
		{ // Mismatching shapes
			p: MustNewType("foo", "struct{ F int; G int }"),
			q: MustNewType("foo", "struct{ F int }"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s+%s", test.p, test.q), func(t *testing.T) {
			m := make(TypeInferenceMap)
			if err := m.Infer(test.p, test.q); err == nil {
				t.Errorf("Infer(%s, %s) = nil error", test.p, test.q)
			}
		})
	}
}
