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

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/server/view"
	"github.com/google/shenzhen-go/dev/source"
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
func GeneratePackage(g *model.Graph) (string, error) {
	gp, err := source.GoPath()
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
	if err := g.WriteGoTo(f); err != nil {
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

// Graph handles displaying/editing a graph.
func renderGraph(g *serveGraph, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s graph: %s", r.Method, r.URL)
	q := r.URL.Query()

	g.Lock()
	defer g.Unlock()

	if _, t := q["up"]; t {
		d := filepath.Dir(g.FilePath)
		http.Redirect(w, r, "/"+d, http.StatusFound)
		return
	}
	if _, t := q["go"]; t {
		outputGoSrc(g.Graph, w)
		return
	}
	if _, t := q["rawgo"]; t {
		outputRawGoSrc(g.Graph, w)
		return
	}
	if _, t := q["json"]; t {
		outputJSON(g.Graph, w)
		return
	}
	if _, t := q["build"]; t {
		if err := Build(g.Graph); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error building:\n%v", err)
			return
		}
		u := *r.URL
		u.RawQuery = ""
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}
	if _, t := q["install"]; t {
		if err := Install(g.Graph); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error installing:\n%v", err)
			return
		}
		u := *r.URL
		u.RawQuery = ""
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}
	if _, t := q["run"]; t {
		w.Header().Set("Content-Type", "text/plain")
		if err := Run(g.Graph, w, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error building or running:\n%v", err)
		}
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unsupported verb %q", r.Method)
		return
	}

	view.Graph(w, g.Graph)
}

func outputGoSrc(g *model.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/plain")
	if err := g.WriteGoTo(w); err != nil {
		log.Printf("Could not render to Go: %v", err)
		http.Error(w, "Could not render to Go", http.StatusInternalServerError)
	}
}

func outputRawGoSrc(g *model.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/plain")
	if err := g.WriteRawGoTo(w); err != nil {
		log.Printf("Could not render to Go: %v", err)
		http.Error(w, "Could not render to Go", http.StatusInternalServerError)
	}
}

func outputJSON(g *model.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "application/json")
	if err := g.WriteJSONTo(w); err != nil {
		log.Printf("Could not encode JSON: %v", err)
		http.Error(w, "Could not encode JSON", http.StatusInternalServerError)
		return
	}
}
