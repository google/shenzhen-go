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
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/zserge/webview"
	"google.golang.org/grpc"

	pb "github.com/google/shenzhen-go/dev/proto/go"
	"github.com/google/shenzhen-go/dev/server"
	"github.com/google/shenzhen-go/dev/server/view"
)

const pingMsg = "Pong!"

var (
	uiAddr            = flag.String("ui_addr", "localhost:0", "Address to bind UI server to")
	useDefaultBrowser = flag.Bool("use_browser", true, "Load in the system's default web browser instead of the inbuilt webview")
)

func systemOpen(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Run()
	case "linux":
		// Seems to work on Linux Mint. YMMV.
		return exec.Command("xdg-open", url).Run()
	case "windows":
		return exec.Command("cmd.exe", "/C", "start", url).Run()
	default:
		fmt.Printf("Ready to open %s\n", url)
		return nil
	}
}

func webviewOpen(url string) error {
	return webview.Open("SHENZHEN GO", url, 1152, 720, true)
}

func isUp(base string) bool {
	resp, err := http.Get(base + "ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return string(msg) == pingMsg
}

func openWhenUp(addr net.Addr, useBrowser bool) {
	base := fmt.Sprintf(`http://%s/`, addr)
	try := time.NewTicker(100 * time.Millisecond)
	defer try.Stop()
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
checkLoop:
	for {
		select {
		case <-try.C:
			if isUp(base) {
				break checkLoop
			}
		case <-timeout.C:
			fmt.Fprintf(os.Stderr, "Couldn't find server after 5 seconds, giving up")
			os.Exit(1)
		}
	}
	open := systemOpen
	if !useBrowser {
		open = webviewOpen
	}
	if err := open(base); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't automatically open: %v\n", err)
		fmt.Printf("Ready to open %s\n", base)
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(pingMsg)) })
	http.Handle("/favicon.ico", view.Favicon)
	http.Handle("/.static/", http.StripPrefix("/.static/", view.Static))

	gs := grpc.NewServer()
	pb.RegisterShenzhenGoServer(gs, server.S)
	ws := grpcweb.WrapServer(gs)
	http.Handle("/.api/", http.StripPrefix("/.api/", ws))

	// Finally, all unknown paths are assumed to be files.
	http.Handle("/", server.S)

	ln, err := net.Listen("tcp", *uiAddr)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}
	defer ln.Close()

	wait := make(chan struct{})
	go func() {
		if err := http.Serve(ln, nil); err != nil {
			log.Fatalf("http.Serve failed: %v", err)
		}
		close(wait)
	}()

	// As soon as we're serving, launch "open" which should launch a browser,
	// or ask the user to do so.
	// This must be called from the main thread to avoid
	// https://github.com/zserge/webview/issues/29.
	openWhenUp(ln.Addr(), *useDefaultBrowser)

	// Job done.
	<-wait
}
