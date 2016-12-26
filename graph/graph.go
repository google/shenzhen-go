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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	PackageName string              `json:"package_name"`
	PackagePath string              `json:"package_path"`
	Imports     []string            `json:"imports"`
	Nodes       map[string]*Node    `json:"nodes"`
	Channels    map[string]*Channel `json:"channels"`
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
	return goimports(w, buf)
}

func (g *Graph) saveGoSrc() error {
	gopath, ok := os.LookupEnv("GOPATH")
	if !ok || gopath == "" {
		return errors.New("cannot use $GOPATH; empty or undefined")
	}
	pp := filepath.Join(gopath, "src", g.PackagePath)
	if err := os.Mkdir(pp, os.FileMode(0755)); err != nil {
		log.Printf("Could not make path %q, continuing: %v", pp, err)
	}
	mp := filepath.Join(pp, "main.go")
	f, err := os.Create(mp)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := g.WriteGoTo(f); err != nil {
		return err
	}
	return f.Close()
}

func (g *Graph) build() error {
	return exec.Command(`go`, `build`, g.PackagePath).Run()
}

// SaveBuildAndRun saves the project as Go source code, builds it, and runs it.
func (g *Graph) SaveBuildAndRun() error {
	if err := g.saveGoSrc(); err != nil {
		return err
	}
	if err := g.build(); err != nil {
		return err
	}
	// TODO: Be less lazy about the output binary path
	return open("./" + g.PackageName)
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
