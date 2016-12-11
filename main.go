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

// The shenzhen-go binary serves a visual Go environment.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

const pingMsg = "Pong!"

var (
	serveAddr = flag.String("addr", "::1", "Address to bind server to")
	servePort = flag.Int("port", 8088, "Port to serve from")

	identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
)

func open(args ...string) error {
	cmd := exec.Command(`open`, args...)
	return cmd.Run()
}

func pipeThru(dst io.Writer, cmd *exec.Cmd, src io.Reader) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(stdin, src); err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	if _, err := io.Copy(dst, stdout); err != nil {
		return err
	}
	return cmd.Wait()
}

func dotToSVG(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`dot`, `-Tsvg`), src)
}

func gofmt(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`gofmt`), src)
}

func goimports(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`goimports`), src)
}

func openWhenUp(addr string) {
	base := fmt.Sprintf("http://%s/", addr)
	t := time.NewTicker(100 * time.Millisecond)
	for range t.C {
		resp, err := http.Get(base + "ping")
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		if string(msg) != pingMsg {
			continue
		}
		open(base)
		t.Stop()
		return
	}
}

func main() {
	flag.Parse()
	addr := net.JoinHostPort(*serveAddr, strconv.Itoa(*servePort))

	http.HandleFunc("/channel/", exampleGraph.handleChannelRequest)
	http.HandleFunc("/node/", exampleGraph.handleNodeRequest)
	http.HandleFunc("/", exampleGraph.handleRootRequest)
	http.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, pingMsg)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		// TODO: serve a favicon properly
		http.Redirect(w, r, "http://golang.org/favicon.ico", http.StatusFound)
	})

	// As soon as we're serving, launch a web browser.
	// Generally expected to work on macOS...
	go openWhenUp(addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
