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
	"regexp"
	"testing"

	"gopkg.in/d4l3k/messagediff.v1"
)

func ExampleTypePattern() {
	mp := NewTypePattern("map[$K]$V")
	types, _ := mp.Infer("map[string]int")
	fmt.Printf("$K = %s, $V = %s", types["$K"], types["$V"])
	// Output: $K = string, $V = int
}

func TestNewTypePattern(t *testing.T) {
	tests := []struct {
		spec string
		want *TypePattern
	}{
		{
			spec: "$T",
			want: &TypePattern{
				spec:   "$T",
				params: []string{"$T"},
				re:     regexp.MustCompile(`^(.+?)$`),
			},
		},
		{
			spec: "$barnacle777",
			want: &TypePattern{
				spec:   "$barnacle777",
				params: []string{"$barnacle777"},
				re:     regexp.MustCompile(`^(.+?)$`),
			},
		},
		{
			spec: "[]$T",
			want: &TypePattern{
				spec:   "[]$T",
				params: []string{"$T"},
				re:     regexp.MustCompile(`^\[\](.+?)$`),
			},
		},
		{
			spec: "map[$K]$V",
			want: &TypePattern{
				spec:   "map[$K]$V",
				params: []string{"$K", "$V"},
				re:     regexp.MustCompile(`^map\[(.+?)\](.+?)$`),
			},
		},
		{
			spec: "map[$T]$T",
			want: &TypePattern{
				spec:   "map[$T]$T",
				params: []string{"$T", "$T"},
				re:     regexp.MustCompile(`^map\[(.+?)\](.+?)$`),
			},
		},
		{
			spec: "struct { F $T }",
			want: &TypePattern{
				spec:   "struct { F $T }",
				params: []string{"$T"},
				re:     regexp.MustCompile(`^struct \{ F (.+?) \}$`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.spec, func(t *testing.T) {
			got := NewTypePattern(test.spec)
			want := test.want
			if got.spec != test.spec {
				t.Errorf("NewTypePattern(%q).spec = %q, want %q", test.spec, got.spec, want.spec)
			}
			if got.re.String() != want.re.String() {
				t.Errorf("NewTypePattern(%q).re = %q, want %q", test.spec, got.re, want.re)
			}
			if diff, equal := messagediff.PrettyDiff(got.params, want.params); !equal {
				t.Errorf("NewTypePattern(%q).params diff: %s", test.spec, diff)
			}
		})
	}
}

func TestTypePatternInfer(t *testing.T) {
	type subtest struct {
		input string
		want  map[string]string
	}
	tests := []struct {
		spec  string
		tests []subtest
	}{
		{
			spec: "$T",
			tests: []subtest{
				{input: "int", want: map[string]string{"$T": "int"}},
				{input: "string", want: map[string]string{"$T": "string"}},
				{input: "[]int", want: map[string]string{"$T": "[]int"}},
				{input: "[]string", want: map[string]string{"$T": "[]string"}},
				{input: "map[string]int", want: map[string]string{"$T": "map[string]int"}},
				{input: "struct{}", want: map[string]string{"$T": "struct{}"}},
			},
		},
		{
			spec: "[]$T",
			tests: []subtest{
				{input: "[]int", want: map[string]string{"$T": "int"}},
				{input: "[]string", want: map[string]string{"$T": "string"}},
				{input: "[][]int", want: map[string]string{"$T": "[]int"}},
				{input: "[]struct{}", want: map[string]string{"$T": "struct{}"}},
				{input: "[]map[string]int", want: map[string]string{"$T": "map[string]int"}},
			},
		},
		{
			spec: "map[$K]$V",
			tests: []subtest{
				{input: "map[int]int", want: map[string]string{"$K": "int", "$V": "int"}},
				{input: "map[string]string", want: map[string]string{"$K": "string", "$V": "string"}},
				{input: "map[string]int", want: map[string]string{"$K": "string", "$V": "int"}},
				{input: "map[int]string", want: map[string]string{"$K": "int", "$V": "string"}},
				{input: "map[interface{}]interface{}", want: map[string]string{"$K": "interface{}", "$V": "interface{}"}},
				{input: "map[struct{}]map[string]string", want: map[string]string{"$K": "struct{}", "$V": "map[string]string"}},
			},
		},
		{
			spec: "map[$T]$T",
			tests: []subtest{
				{input: "map[int]int", want: map[string]string{"$T": "int"}},
				{input: "map[string]string", want: map[string]string{"$T": "string"}},
				{input: "map[interface{}]interface{}", want: map[string]string{"$T": "interface{}"}},
			},
		},
		{
			spec: "struct { F $T }",
			tests: []subtest{
				{input: "struct { F int }", want: map[string]string{"$T": "int"}},
				{input: "struct { F string }", want: map[string]string{"$T": "string"}},
				{input: "struct { F []int }", want: map[string]string{"$T": "[]int"}},
				{input: "struct { F struct{} }", want: map[string]string{"$T": "struct{}"}},
				{input: "struct { F map[string]int }", want: map[string]string{"$T": "map[string]int"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.spec, func(t *testing.T) {
			p := NewTypePattern(test.spec)
			for _, st := range test.tests {
				t.Run(st.input, func(t *testing.T) {
					got, err := p.Infer(st.input)
					if err != nil {
						t.Fatalf("Infer(%q) = error %v", st.input, err)
					}
					want := st.want
					if diff, equal := messagediff.PrettyDiff(got, want); !equal {
						t.Errorf("Infer(%q) diff: %s", st.input, diff)
					}
				})
			}
		})
	}
}

func TestTypePatternInferError(t *testing.T) {
	tests := []struct {
		spec  string
		tests []string
	}{
		{
			spec:  "[]$T",
			tests: []string{"int", "string", "map[int]string", "struct{}", "interface{}"},
		},
		{
			spec:  "map[$K]$V",
			tests: []string{"int", "string", "[]string", "struct{}", "interface{}"},
		},
		{
			spec: "map[$T]$T",
			tests: []string{"int", "string", "[]string", "struct{}", "interface{}",
				"map[string]int", "map[int]string", "map[string]map[string]string"},
		},
		{
			spec: "struct { F $T }",
			tests: []string{"int", "string", "[]string", "struct{}", "interface{}",
				"struct { F string\nG string }", " struct { G string }"},
		},
	}

	for _, test := range tests {
		t.Run(test.spec, func(t *testing.T) {
			p := NewTypePattern(test.spec)
			for _, input := range test.tests {
				t.Run(input, func(t *testing.T) {
					_, err := p.Infer(input)
					if err == nil {
						t.Fatalf("Infer(%q) = nil error", input)
					}
				})
			}
		})
	}
}
