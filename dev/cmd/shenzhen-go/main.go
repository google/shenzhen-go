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
	"errors"
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
	"google.golang.org/grpc"

	_ "github.com/google/shenzhen-go/dev/parts"
	pb "github.com/google/shenzhen-go/dev/proto/go"
	"github.com/google/shenzhen-go/dev/server"
	"github.com/google/shenzhen-go/dev/server/view"
)

const pingMsg = "Pong!"

var (
	uiAddr = flag.String("ui_addr", "localhost:0", "`address` to bind UI server to")

	// Set up by webview.go
	useDefaultBrowser *bool
	webviewOpen       func(string) error

	viewParams view.Params
)

func init() {
	flag.StringVar(&viewParams.AceTheme, "ace_theme", "chrome", "name of the ace theme (not including ace/theme/ prefix)")
	flag.StringVar(&viewParams.CSSTheme, "css_theme", "default", "name of theme css file to use (name as in theme-${name}.css)")
}

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

func isUp(url string) bool {
	resp, err := http.Get(url)
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

func waitForUp(addr net.Addr) error {
	url := fmt.Sprintf(`http://%s/ping`, addr)
	try := time.NewTicker(100 * time.Millisecond)
	defer try.Stop()
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case <-try.C:
			if isUp(url) {
				return nil
			}
		case <-timeout.C:
			return errors.New("timed out waiting 5s for server")
		}
	}
}

func open(addr net.Addr, path string, useBrowser bool) {
	url := fmt.Sprintf(`http://%s/%s`, addr, path)
	opener := systemOpen
	if !useBrowser && webviewOpen != nil {
		opener = webviewOpen
	}
	if err := opener(url); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't automatically open: %v\n", err)
		fmt.Printf("Ready to open %s\n", url)
	}
}

func serve(addr chan<- net.Addr) error {
	ln, err := net.Listen("tcp", *uiAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %v", err)
	}
	defer ln.Close()

	if addr != nil {
		addr <- ln.Addr()
	}

	s := server.New(viewParams)

	http.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(pingMsg)) })
	http.Handle("/favicon.ico", view.Favicon)
	http.Handle("/.static/", http.StripPrefix("/.static/", view.Static))
	http.Handle("/", s)

	gs := grpc.NewServer()
	pb.RegisterShenzhenGoServer(gs, s)
	ws := grpcweb.WrapServer(gs,
		grpcweb.WithWebsockets(true),
	)

	svr := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Addr:              ln.Addr().String(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case ws.IsGrpcWebRequest(r):
				ws.HandleGrpcWebRequest(w, r)
			case ws.IsGrpcWebSocketRequest(r):
				ws.HandleGrpcWebsocketRequest(w, r)
			default:
				http.DefaultServeMux.ServeHTTP(w, r)
			}
		}),
	}
	if err := svr.Serve(ln); err != nil {
		return fmt.Errorf("http.Serve: %v", err)
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Shenzhen Go is a tool for managing and editing Shenzhen Go source code.
	
Usage:

  %s [command] [files]
  
The (optional) commands are:
  
  build     generate and build Go packages
  edit      launch a Shenzhen Go server and open the editor interface
  generate  generate Go packages
  install   generate and install Go packages
  run       generate Go package and run binaries
  serve     launch a Shenzhen Go server
  
"edit" is the default command.

Flags:

`, os.Args[0])
	// TODO: Add per-command help. `shenzhen-go help [command]`

	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	openUI := true
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "build":
			log.Fatalf("TODO: build is not yet implemented")
		case "edit":
			args = args[1:]
		case "generate":
			log.Fatalf("TODO: generate is not yet implemented")
		case "help":
			usage()
			return
		case "install":
			log.Fatalf("TODO: install is not yet implemented")
		case "run":
			log.Fatalf("TODO: run is not yet implemented")
		case "serve":
			if len(args) > 1 {
				log.Print(`Note: extra arguments to "serve" command are ignored`)
			}
			openUI = false
		default:
			// Edit, but every arg is a file.
		}
	} else {
		// Opens the browser at the root.
		args = []string{""}
	}

	adch := make(chan net.Addr)
	wait := make(chan struct{})
	go func() {
		serve(adch)
		close(wait)
	}()

	// Wait until properly serving.
	addr := <-adch
	if err := waitForUp(addr); err != nil {
		log.Fatalf("Couldn't reach server: %v", err)
	}
	log.Printf("Serving HTTP on %v", addr)

	if openUI {
		for _, a := range args {
			// Launch "open" which should launch a browser,
			// or ask the user to do so.
			// This must be called from the main thread to avoid
			// https://github.com/zserge/webview/issues/29.
			ub := useDefaultBrowser != nil && *useDefaultBrowser
			open(addr, a, ub)
		}
	}

	// Job done.
	<-wait
}
