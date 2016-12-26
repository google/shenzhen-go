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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"shenzhen-go/view"
)

const pingMsg = "Pong!"

var (
	serveAddr = flag.String("addr", "::1", "Address to bind server to")
	servePort = flag.Int("port", 8088, "Port to serve from")
)

func open(args ...string) error {
	cmd := exec.Command(`open`, args...)
	return cmd.Run()
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

	r := mux.NewRouter()
	r.Handle("/channel/{chan}", (*view.ChannelHandler)(exampleGraph))
	r.Handle("/node/{node}", (*view.NodeHandler)(exampleGraph))
	r.Handle("/browse/", http.StripPrefix("/browse/", view.DirBrowser{}))
	r.Handle("/", (*view.GraphHandler)(exampleGraph))

	r.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, pingMsg)
	})
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		// TODO: serve a favicon properly
		http.Redirect(w, r, "http://golang.org/favicon.ico", http.StatusFound)
	})

	http.Handle("/", r)

	// As soon as we're serving, launch a web browser.
	// Generally expected to work on macOS...
	go openWhenUp(addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
