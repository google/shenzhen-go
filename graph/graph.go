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

// Package graph manages programs stored as graphs.
package graph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Channel models a channel. It can be marshalled and unmarshalled to JSON sensibly.
type Channel struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Cap  int    `json:"cap"`
}

// Graph describes a Go program as a graph. It can be marshalled and unmarshalled to JSON sensibly.
type Graph struct {
	SourcePath  string              `json:"-"` // path to the JSON source.
	Name        string              `json:"name"`
	PackagePath string              `json:"package_path"`
	Imports     []string            `json:"imports"`
	IsCommand   bool                `json:"is_command"`
	Nodes       map[string]*Node    `json:"nodes"`
	Channels    map[string]*Channel `json:"channels"`
}

// New returns a new empty graph associated with a file path.
func New(srcPath string) *Graph {
	return &Graph{
		SourcePath: srcPath,
		Nodes:      make(map[string]*Node),
		Channels:   make(map[string]*Channel),
	}
}

// PackageName extracts the name of the package from the package path ("full" package name).
func (g *Graph) PackageName() string {
	i := strings.LastIndex(g.PackagePath, "/")
	if i < 0 {
		return g.PackagePath
	}
	return g.PackagePath[i+1:]
}

// LoadJSON loads a JSON-encoded Graph from an io.Reader.
func LoadJSON(r io.Reader, sourcePath string) (*Graph, error) {
	dec := json.NewDecoder(r)
	var g Graph
	if err := dec.Decode(&g); err != nil {
		return nil, err
	}
	g.SourcePath = sourcePath
	return &g, nil
}

// LoadJSONFile loads a JSON-encoded Graph from a file at a given path.
func LoadJSONFile(path string) (*Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadJSON(f, path)
}

// WriteJSONTo writes nicely-formatted JSON to the given Writer.
func (g *Graph) WriteJSONTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t") // For diffability!
	return enc.Encode(g)
}

// SaveJSONFile saves the JSON-encoded Graph to the SourcePath.
func (g *Graph) SaveJSONFile() error {
	f, err := ioutil.TempFile(filepath.Dir(g.SourcePath), filepath.Base(g.SourcePath))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := g.WriteJSONTo(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), g.SourcePath)
}

// WriteDotTo writes the Dot language view of the graph to the io.Writer.
func (g *Graph) WriteDotTo(dst io.Writer) error { return dotTemplate.Execute(dst, g) }

// WriteGoTo writes the Go language view of the graph to the io.Writer.
func (g *Graph) WriteGoTo(w io.Writer) error {
	buf := &bytes.Buffer{}
	if err := goTemplate.Execute(buf, g); err != nil {
		return err
	}
	return gofmt(w, buf)
}

// WriteGoRunnerTo outputs a simple main package and function for calling Run on
// a generated Go package.
func (g *Graph) WriteGoRunnerTo(w io.Writer) error {
	return goRunnerTemplate.Execute(w, g)
}

// GeneratePackage writes the Go view of the graph to a file called generated.go in
// ${GOPATH}/src/${g.PackagePath}/, returning the full path.
func (g *Graph) GeneratePackage() (string, error) {
	gopath, ok := os.LookupEnv("GOPATH")
	if !ok || gopath == "" {
		return "", errors.New("cannot use $GOPATH; empty or undefined")
	}
	pp := filepath.Join(gopath, "src", g.PackagePath)
	if err := os.Mkdir(pp, os.FileMode(0755)); err != nil {
		log.Printf("Could not make path %q, continuing: %v", pp, err)
	}
	mp := filepath.Join(pp, "generated.go")
	f, err := os.Create(mp)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := g.WriteGoTo(f); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return mp, nil
}

// Build saves the graph as Go source code and tries to build it.
func (g *Graph) Build() error {
	if _, err := g.GeneratePackage(); err != nil {
		return err
	}
	o, err := exec.Command(`go`, `build`, g.PackagePath).CombinedOutput()
	if err != nil {
		// TODO: better error type
		return fmt.Errorf("go build: %v:\n%s", err, o)
	}
	return nil
}

func (g *Graph) writeTempRunner() (string, error) {
	fn := filepath.Join(os.TempDir(), fmt.Sprintf("shenzhen-go-runner.%s.go", g.PackageName()))
	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := g.WriteGoRunnerTo(f); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return fn, nil
}

// Run saves the graph as Go source code, creates a temporary runner, and tries to run it.
// The stdout and stderr pipes are copied to the given io.Writers.
func (g *Graph) Run(stdout, stderr io.Writer) error {
	// Don't have to explicitly build, but must at least have the file ready
	// so that go run can build it.
	gp, err := g.GeneratePackage()
	if err != nil {
		return err
	}

	// TODO: Support stdin?

	if !g.IsCommand {
		// Since it's a library which needs Run to be called,
		// generate and run the temporary runner.
		p, err := g.writeTempRunner()
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

// DeclaredChannels returns the given channels which exist in g.Channels.
func (g *Graph) DeclaredChannels(chans []string) []string {
	r := make([]string, 0, len(chans))
	for _, d := range chans {
		if _, found := g.Channels[d]; found {
			r = append(r, d)
		}
	}
	return r
}
