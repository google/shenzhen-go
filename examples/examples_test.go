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

package examples

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/shenzhen-go/model"
	_ "github.com/google/shenzhen-go/parts"
	"github.com/google/shenzhen-go/source"
)

func TestLoadAndGoExamples(t *testing.T) {
	gp, err := source.GoPath()
	if err != nil {
		t.Fatalf("GoPath() = error %v", err)
	}
	glob := filepath.Join(gp, "src", "github.com/google/shenzhen-go/examples/*.szgo")
	exs, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("Glob(%s) = error %v", glob, err)
	}
	if len(exs) == 0 {
		t.Fatalf("Glob(%s) = %v, want at least some files", glob, exs)
	}
	for _, ex := range exs {
		t.Run(ex, func(t *testing.T) {
			f, err := os.Open(ex)
			if err != nil {
				t.Fatalf("Open(%s) = error %v", ex, err)
			}
			defer f.Close()
			g, err := model.LoadJSON(f, ex, filepath.Base(ex))
			if err != nil {
				t.Fatalf("LoadJSON() = error %v", err)
			}
			if _, err := g.Go(); err != nil {
				t.Fatalf("Go() = error %v", err)
			}
		})
	}
}
