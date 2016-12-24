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
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Channel models a channel.
type Channel struct {
	Name, Type string
	Cap        int
}

// Graph describes a Go program as a graph.
type Graph struct {
	Name        string
	PackageName string
	PackagePath string
	Imports     []string
	Nodes       map[string]*Node
	Channels    map[string]*Channel
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
