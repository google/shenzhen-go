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

type TF struct {
	*testing.T
}

func (t TF) MustNewType(scope, spec string) *Type {
	typ, err := NewType(scope, spec)
	if err != nil {
		t.Fatalf("NewType(%s, %s) = error %v", scope, spec, err)
	}
	return typ
}

func TestTypeRefine(t *testing.T) {
	tf := TF{t}
	tests := []struct {
		base *Type
		in   map[TypeParam]*Type
		want string
	}{
		{
			base: tf.MustNewType("foo", "$T"),
			in: map[TypeParam]*Type{
				{"foo", "$T"}: tf.MustNewType("whatev", "int"),
			},
			want: "int",
		},
		{
			base: tf.MustNewType("bar", "$U"),
			in: map[TypeParam]*Type{
				{"bar", "$U"}: tf.MustNewType("whatev", "string"),
			},
			want: "string",
		},
		{
			base: tf.MustNewType("foo", "$T"),
			in: map[TypeParam]*Type{
				{"bar", "$U"}: tf.MustNewType("whatev", "string"),
			},
			want: "$T",
		},
		{
			base: tf.MustNewType("foo", "map[$K]$V"),
			in: map[TypeParam]*Type{
				{"foo", "$K"}: tf.MustNewType("whatev", "string"),
				{"foo", "$V"}: tf.MustNewType("whatev", "int"),
			},
			want: "map[string]int",
		},
	}

	for _, test := range tests {
		t.Run(test.base.String(), func(t *testing.T) {
			if err := test.base.Refine(test.in); err != nil {
				t.Fatalf("base(%s).Refine(%v) = error %v", test.base, test.in, err)
			}
			if got, want := test.base.String(), test.want; got != want {
				t.Errorf("base = %s, want %s", got, want)
			}
		})
	}
}

func TestTypeLithify(t *testing.T) {
	tf := TF{t}
	tests := []struct {
		base *Type
		lith *Type
		want string
	}{
		{
			base: tf.MustNewType("foo", "$T"),
			lith: tf.MustNewType("irrelevant", "int"),
			want: "int",
		},
		{
			base: tf.MustNewType("foo", "*$T"),
			lith: tf.MustNewType("irrelevant", "int"),
			want: "*int",
		},
		{
			base: tf.MustNewType("foo", "map[$K]$V"),
			lith: tf.MustNewType("irrelevant", "int"),
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
