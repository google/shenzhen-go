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

// Package controller manages everything.
package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/source"
)

// GuessPackagePath attempts to find a sensible package path.
func GuessPackagePath(srcPath string) (string, error) {
	gp, err := gopath()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(srcPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(gp+"/src", abs)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(rel, filepath.Ext(rel)), nil
}

// Definitions returns the imports and channel var blocks from the Go program.
// This is useful for advanced parsing and typechecking.
func Definitions(g *model.Graph) string {
	b := new(bytes.Buffer)
	goDefinitionsTemplate.Execute(b, g)
	return b.String()
}

// AllImports combines all desired imports into one slice.
// It doesn't fix conflicting names, but dedupes any whole lines.
// TODO: Put nodes in separate files to solve all import issues.
func AllImports(g *model.Graph) []string {
	m := source.NewStringSet()
	m.Add(`"sync"`)
	for _, n := range g.Nodes {
		for _, i := range n.Part.Imports() {
			m.Add(i)
		}
	}
	return m.Slice()
}

// LoadJSONFile loads a JSON-encoded Graph from a file at a given path.
func LoadJSONFile(path string) (*model.Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return model.LoadJSON(f, path)
}

// WriteJSONTo writes nicely-formatted JSON to the given Writer.
func WriteJSONTo(w io.Writer, g *model.Graph) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t") // For diffability!
	return enc.Encode(g)
}

// SaveJSONFile saves the JSON-encoded Graph to the SourcePath.
func SaveJSONFile(g *model.Graph) error {
	f, err := ioutil.TempFile(filepath.Dir(g.SourcePath), filepath.Base(g.SourcePath))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := WriteJSONTo(f, g); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), g.SourcePath)
}

// WriteGoTo writes the Go language view of the graph to the io.Writer.
func WriteGoTo(w io.Writer, g *model.Graph) error {
	buf := &bytes.Buffer{}
	if err := WriteRawGoTo(buf, g); err != nil {
		return err
	}
	return gofmt(w, buf)
}

// WriteRawGoTo writes the Go language view of the graph to the io.Writer, without formatting.
func WriteRawGoTo(w io.Writer, g *model.Graph) error {
	return goTemplate.Execute(w, g)
}

// GeneratePackage writes the Go view of the graph to a file called generated.go in
// ${GOPATH}/src/${g.PackagePath}/, returning the full path.
func GeneratePackage(g *model.Graph) (string, error) {
	gp, err := gopath()
	if err != nil {
		return "", err
	}
	pp := filepath.Join(gp, "src", g.PackagePath)
	if err := os.Mkdir(pp, os.FileMode(0755)); err != nil {
		log.Printf("Could not make path %q, continuing: %v", pp, err)
	}
	mp := filepath.Join(pp, "generated.go")
	f, err := os.Create(mp)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := WriteGoTo(f, g); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return mp, nil
}

// Build saves the graph as Go source code and tries to "go build" it.
func Build(g *model.Graph) error {
	if _, err := GeneratePackage(g); err != nil {
		return err
	}
	o, err := exec.Command(`go`, `build`, g.PackagePath).CombinedOutput()
	if err != nil {
		// TODO: better error type
		return fmt.Errorf("go build: %v:\n%s", err, o)
	}
	return nil
}

// Install saves the graph as Go source code and tries to "go install" it.
func Install(g *model.Graph) error {
	if _, err := GeneratePackage(g); err != nil {
		return err
	}
	o, err := exec.Command(`go`, `install`, g.PackagePath).CombinedOutput()
	if err != nil {
		// TODO: better error type
		return fmt.Errorf("go install: %v:\n%s", err, o)
	}
	return nil
}

func writeTempRunner(g *model.Graph) (string, error) {
	fn := filepath.Join(os.TempDir(), fmt.Sprintf("shenzhen-go-runner.%s.go", g.PackageName()))
	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := goRunnerTemplate.Execute(f, g); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return fn, nil
}

// Run saves the graph as Go source code, creates a temporary runner, and tries to run it.
// The stdout and stderr pipes are copied to the given io.Writers.
func Run(g *model.Graph, stdout, stderr io.Writer) error {
	// Don't have to explicitly build, but must at least have the file ready
	// so that go run can build it.
	gp, err := GeneratePackage(g)
	if err != nil {
		return err
	}

	// TODO: Support stdin?

	if !g.IsCommand {
		// Since it's a library which needs Run to be called,
		// generate and run the temporary runner.
		p, err := writeTempRunner(g)
		if err != nil {
			return err
		}
		gp = p
	}
	cmd := exec.Command(`go`, `run`, gp)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	e, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go io.Copy(stdout, o)
	go io.Copy(stderr, e)
	return cmd.Wait()
}
