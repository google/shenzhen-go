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

package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/server/view"
	"github.com/google/shenzhen-go/source"
)

var identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

// GuessPackagePath attempts to find a sensible package path.
func GuessPackagePath(srcPath string) (string, error) {
	gp, err := source.GoPath()
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

// SaveJSONFile saves the JSON-encoded Graph to the SourcePath.
func SaveJSONFile(g *model.Graph) error {
	f, err := ioutil.TempFile(filepath.Dir(g.FilePath), filepath.Base(g.FilePath))
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
	return os.Rename(f.Name(), g.FilePath)
}

// GeneratePackage writes the Go view of the graph to a file called generated.go in
// ${GOPATH}/src/${g.PackagePath}/, returning the full path.
// Messages from the generation process will be written to out.
func GeneratePackage(out io.Writer, g *model.Graph) (string, error) {
	fmt.Fprintln(out, "[GeneratePackage]")
	gp, err := source.GoPath()
	if err != nil {
		fmt.Fprintf(out, "source.GoPath() = %v\n(GeneratePackage failed)\n", err)
		return "", err
	}
	pp := filepath.Join(gp, "src", g.PackagePath)
	if err := os.MkdirAll(pp, os.FileMode(0755)); err != nil {
		fmt.Fprintf(out, "os.MkdirAll(pp, 0755) = %v)\n", err)
		return "", err
	}
	mp := filepath.Join(pp, "generated.go")
	f, err := os.Create(mp)
	if err != nil {
		fmt.Fprintf(out, "os.Create(mp) = %v\n(GeneratePackage failed)\n", err)
		return "", err
	}
	defer f.Close()
	if err := g.WriteGoTo(f); err != nil {
		fmt.Fprintf(out, "g.WriteGoTo(f) = %v\n(GeneratePackage failed)\n", err)
		return "", err
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(out, "f.Close() = %v\n(GeneratePackage failed)\n", err)
		return "", err
	}
	fmt.Fprintln(out, "(GeneratePackage succeeded)")
	return mp, nil
}

// GenerateRunner generates a `go run`-able; either the output package itself,
// or the package together with a temporary runner, returning the full path to
// the runnable path. Messages from the generation process will be written to out.
func GenerateRunner(out io.Writer, g *model.Graph) (string, error) {
	gp, err := GeneratePackage(out, g)
	if err != nil {
		return "", err
	}
	if g.IsCommand {
		return gp, nil
	}
	fmt.Fprintln(out, "[GenerateRunner]")
	path, err := writeTempRunner(g)
	if err != nil {
		fmt.Fprintf(out, "writeTempRunner(g) = %v\n(GenerateRunner failed)\n", err)
		return "", err
	}
	fmt.Fprintln(out, "(GenerateRunner succeeded)")
	return path, nil
}

func runCmd(out io.Writer, cmd *exec.Cmd) error {
	fmt.Fprintf(out, "%v\n", cmd.Args)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(out, "(process %v)\n", err)
		return err
	}
	fmt.Fprintln(out, "(process succeeded)")
	return nil
}

// Build saves the graph as Go source code and tries to "go build" it.
// Console output from the command (*not* the compiled program) is written to out.
func Build(out io.Writer, g *model.Graph) error {
	if _, err := GeneratePackage(out, g); err != nil {
		return err
	}
	return runCmd(out, exec.Command(`go`, `build`, g.PackagePath))
}

// Install saves the graph as Go source code and tries to "go install" it.
// Console output from the command (*not* the compiled program) is written to out.
func Install(out io.Writer, g *model.Graph) error {
	if _, err := GeneratePackage(out, g); err != nil {
		return err
	}
	return runCmd(out, exec.Command(`go`, `install`, g.PackagePath))
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

// Graph handles displaying/editing a graph.
func renderGraph(g *serveGraph, w http.ResponseWriter, r *http.Request, uiParams *view.Params) {
	log.Printf("%s graph: %s", r.Method, r.URL)
	q := r.URL.Query()

	g.Lock()
	defer g.Unlock()

	if _, t := q["up"]; t {
		d := filepath.Dir(g.FilePath)
		http.Redirect(w, r, "/"+d, http.StatusFound)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unsupported method %q", r.Method)
		return
	}

	view.Graph(w, g.Graph, uiParams)
}
