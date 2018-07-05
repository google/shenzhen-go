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

func TestTypeRefine(t *testing.T) {
	tests := []struct {
		base *Type
		in   map[TypeParam]*Type
		want string
	}{
		{
			base: MustNewType("foo", "struct{}"),
			in: map[TypeParam]*Type{
				{"foo", "$T"}: MustNewType("", "int"),
			},
			want: "struct{}",
		},
		{
			base: MustNewType("foo", "$T"),
			in: map[TypeParam]*Type{
				{"foo", "$T"}: MustNewType("", "int"),
			},
			want: "int",
		},
		{
			base: MustNewType("bar", "$U"),
			in: map[TypeParam]*Type{
				{"bar", "$U"}: MustNewType("", "string"),
			},
			want: "string",
		},
		{
			base: MustNewType("foo", "$T"),
			in: map[TypeParam]*Type{
				{"bar", "$U"}: MustNewType("", "string"),
			},
			want: "$T",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: map[TypeParam]*Type{
				{"foo", "$K"}: MustNewType("", "string"),
				{"foo", "$V"}: MustNewType("", "int"),
			},
			want: "map[string]int",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: map[TypeParam]*Type{
				{"foo", "$V"}: MustNewType("", "int"),
			},
			want: "map[$K]int",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in: map[TypeParam]*Type{
				{"foo", "$K"}: MustNewType("", "int"),
			},
			want: "map[int]$V",
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

func TestTypeLithify(t *testing.T) {
	tests := []struct {
		base *Type
		lith *Type
		want string
	}{
		{
			base: MustNewType("foo", "$T"),
			lith: MustNewType("", "int"),
			want: "int",
		},
		{
			base: MustNewType("foo", "*$T"),
			lith: MustNewType("", "int"),
			want: "*int",
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			lith: MustNewType("", "int"),
			want: "map[int]int",
		},
	}

	for _, test := range tests {
		t.Run(test.base.String(), func(t *testing.T) {
			if err := test.base.Lithify(test.lith); err != nil {
				t.Fatalf("base(%s).Lithify(%s) = error %v", test.base, test.lith, err)
			}
			if got, want := test.base.String(), test.want; got != want {
				t.Errorf("base = %s, want %s", got, want)
			}
		})
	}
}

func TestTypeInfer(t *testing.T) {
	tests := []struct {
		base *Type
		in   *Type
		want map[TypeParam]string
	}{
		{
			base: MustNewType("foo", "int"),
			in:   MustNewType("", "int"),
			want: map[TypeParam]string{},
		},
		{
			base: MustNewType("foo", "packaged.Type"),
			in:   MustNewType("", "packaged.Type"),
			want: map[TypeParam]string{},
		},
		{
			base: MustNewType("foo", "$T"),
			in:   MustNewType("", "int"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "int",
			},
		},
		{
			base: MustNewType("foo", "*$T"),
			in:   MustNewType("", "*int"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "int",
			},
		},
		{
			base: MustNewType("bar", "[]$T"),
			in:   MustNewType("", "[]string"),
			want: map[TypeParam]string{
				{"bar", "$T"}: "string",
			},
		},
		{
			base: MustNewType("foo", "map[$K]$V"),
			in:   MustNewType("", "map[interface{}]struct{}"),
			want: map[TypeParam]string{
				{"foo", "$K"}: "interface{}",
				{"foo", "$V"}: "struct{}",
			},
		},
		{
			base: MustNewType("foo", "struct{F $T; G $U}"),
			in:   MustNewType("", "struct { F float64; G complex128 }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "float64",
				{"foo", "$U"}: "complex128",
			},
		},
		{
			base: MustNewType("foo", "map[$T]$T"),
			in:   MustNewType("", "map[interface{}]interface{}"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "interface{}",
			},
		},
		{
			base: MustNewType("foo", "$T"),
			in:   MustNewType("bar", "$U"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "$U",
			},
		},
		{
			base: MustNewType("foo", "map[$K]string"),
			in:   MustNewType("bar", "map[int]$V"),
			want: map[TypeParam]string{
				{"foo", "$K"}: "int",
			},
		},
		{
			base: MustNewType("foo", "struct{F $T; G $T}"),
			in:   MustNewType("bar", "struct { F map[$K]$V; G map[string]int }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[string]int",
			},
		},
		{
			base: MustNewType("foo", "struct{F $T; G $T}"),
			in:   MustNewType("bar", "struct { F map[string]int; G map[$K]$V }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[string]int",
			},
		},
		/* {  // Infer,Infer,Refine
			base: MustNewType("foo", "struct{F $T; G $T}"),
			in:   MustNewType("bar", "struct { F map[$K]int; G map[string]$V }"),
			want: map[TypeParam]string{
				{"foo", "$T"}: "map[string]int",
			},
		}, */
	}

	for _, test := range tests {
		t.Run(test.base.String(), func(t *testing.T) {
			inf, err := test.base.Infer(test.in)
			if err != nil {
				t.Fatalf("base(%s).Infer(%s) = error %v", test.base, test.in, err)
			}
			got := make(map[TypeParam]string)
			for param, typ := range inf {
				got[param] = typ.String()
			}
			if diff, equal := messagediff.PrettyDiff(got, test.want); !equal {
				t.Errorf("base.Infer diff\n%s", diff)
			}
		})
	}
}
