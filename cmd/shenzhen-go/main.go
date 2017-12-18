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
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	pb "github.com/google/shenzhen-go/proto"
	"github.com/google/shenzhen-go/server"
	"github.com/google/shenzhen-go/server/view"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

const pingMsg = "Pong!"

var uiAddr = flag.String("ui_addr", "localhost:8088", "Address to bind UI server to")

func open(args ...string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", args...).Run()
	case "linux":
		// TODO: Just guessing, fix later.
		return exec.Command("xdg-open", args...).Run()
	case "windows":
		return exec.Command("start", args...).Run()
	default:
		fmt.Printf("Ready to open %s\n", strings.Join(args, " "))
		return nil
	}
}

// TODO: Implement this better.
func openWhenUp(addr string) {
	base := fmt.Sprintf(`http://%s/`, addr)
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
		if err := open(base); err != nil {
			fmt.Printf("Ready to open %s\n", base)
		}
		t.Stop()
		return
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, pingMsg)
	})

	http.Handle("/favicon.ico", view.Favicon)
	http.Handle("/.static/", http.StripPrefix("/.static/", view.Static))

	gs := grpc.NewServer()
	pb.RegisterShenzhenGoServer(gs, server.S)
	ws := grpcweb.WrapServer(gs)
	http.Handle("/.api/", http.StripPrefix("/.api/", ws))

	// Finally, all unknown paths are assumed to be files.
	http.Handle("/", server.S)

	// As soon as we're serving, launch "open" which should launch a browser,
	// or ask the user to do so.
	go openWhenUp(*uiAddr)

	if err := http.ListenAndServe(*uiAddr, nil); err != nil {
		log.Fatal(err)
	}
}
