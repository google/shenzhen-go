// Copyright 2018 Google Inc.
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

// +build mage

package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// gzips a file, overwriting the original. Does not do a file dance.
func inPlaceGzip(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("inPlaceGzip: got directory %q, want a file", path)
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	o, err := os.Create(path)
	if err != nil {
		return err
	}
	defer o.Close()
	g, err := gzip.NewWriterLevel(o, gzip.BestCompression)
	if err != nil {
		return err
	}
	g.Header.Name = filepath.Base(path)
	g.Header.ModTime = fi.ModTime()
	if _, err := g.Write(b); err != nil {
		return err
	}
	if err := g.Close(); err != nil {
		return err
	}
	return o.Close()
}

// Runs protoc to generate protobuf stubs in proto/{go,js}.
func GenProtoStubs() error {
	return sh.Run("protoc", "-I=proto", "shenzhen-go.proto", "--go_out=plugins=grpc:proto/go", "--gopherjs_out=plugins=grpc:proto/js")
}

// Builds the client into server/view/js/client.js{,.map}.
func BuildClient() error {
	mg.Deps(GenProtoStubs)

	// Per the GopherJS README, supported GOOS values are {linux, darwin}.
	// Force GOOS=linux to support building on Windows, and also
	// help keep the file stable (I dev on both Linux and macOS).
	env := map[string]string{
		"GOOS": "linux",
	}
	if err := sh.RunWith(env, "gopherjs", "build", "-o", "server/view/js/client.js", "github.com/google/shenzhen-go/client"); err != nil {
		return err
	}
	// The server transparently handles gzipped embedded content.
	if err := inPlaceGzip("server/view/js/client.js"); err != nil {
		return err
	}
	return inPlaceGzip("server/view/js/client.js.map")
}

// Embeds static content into static-*.go files in the server/view package.
func Embed() error {
	mg.Deps(BuildClient)

	embed := sh.RunCmd("go", "run", "scripts/embed/embed.go", "-p", "view", "-base", "server/view")

	embeds := [][]string{
		{"-v", "cssResources", "-o", "server/view/static-css.go", "-gzip", "css/*.css"},
		{"-v", "imageResources", "-o", "server/view/static-images.go", "images/*"},
		{"-v", "jsResources", "-o", "server/view/static-js.go", "-gzip", "js/*", "js/*/*"},
		{"-v", "miscResources", "-o", "server/view/static-misc.go", "-gzip", "misc/*"},
		{"-v", "templateResources", "-o", "server/view/static-templates.go", "templates/*.html"},
	}

	for _, args := range embeds {
		if err := embed(args...); err != nil {
			return err
		}
	}
	return nil
}

// Install rebuilds everything and then go-installs.
func Install() error {
	mg.Deps(Embed)
	return sh.Run("go", "install")
}

// GoGetTools uses "go get" to get and update necessary build tools for development.
func GoGetTools() error {
	goGet := sh.RunCmd("go", "get", "-u")

	if err := goGet("github.com/gopherjs/gopherjs"); err != nil {
		return err
	}
	if err := goGet("github.com/golang/protobuf/protoc-gen-go"); err != nil {
		return err
	}
	return goGet("github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs")
}

// Clean removes all of code that can be regenerated.
func Clean() error {
	intfs := []string{
		"proto/go/shenzhen-go.pb.go",
		"proto/js/shenzhen-go.pb.gopherjs.go",
		"server/view/js/client.js",
		"server/view/js/client.js.map",
		"server/view/static-css.go",
		"server/view/static-images.go",
		"server/view/static-js.go",
		"server/view/static-misc.go",
		"server/view/static-templates.go",
	}
	for _, intf := range intfs {
		if err := sh.Rm(intf); err != nil {
			return err
		}
	}
	return nil
}
