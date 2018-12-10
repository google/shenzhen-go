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
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

// Runs protoc to generate protobuf stubs in proto/go.
func GenGoProtoStubs() error {
	mod, err := target.Path("proto/go/shenzhen-go.pb.go", "proto/shenzhen-go.proto")
	if err != nil {
		return err
	}
	if !mod {
		return nil
	}
	return sh.Run("protoc", "-I=proto", "shenzhen-go.proto", "--go_out=plugins=grpc:proto/go")
}

// Runs protoc to generate protobuf stubs in proto/js.
func GenGopherJSProtoStubs() error {
	mod, err := target.Path("proto/js/shenzhen-go.pb.gopherjs.go", "proto/shenzhen-go.proto")
	if err != nil {
		return err
	}
	if !mod {
		return nil
	}
	return sh.Run("protoc", "-I=proto", "shenzhen-go.proto", "--gopherjs_out=plugins=grpc:proto/js")
}

// Builds the client into server/view/js/client.js{,.map}.
func BuildClient() error {
	mg.Deps(GenGopherJSProtoStubs)

	mod, err := target.Dir("server/view/js/client.js", "client", "dom", "parts", "proto")
	if err != nil {
		return err
	}
	if !mod {
		return nil
	}

	// Per the GopherJS README, supported GOOS values are {linux, darwin}.
	// Force GOOS=linux to support building on Windows, and also
	// help keep the file stable (I dev on both Linux and macOS).
	env := map[string]string{
		"GOOS": "linux",
	}
	return sh.RunWith(env, "gopherjs", "build", "-o", "server/view/js/client.js", "github.com/google/shenzhen-go/client")
}

// Embeds static content into static-*.go files in the server/view package.
func Embed() error {
	mg.Deps(BuildClient)

	embed := sh.RunCmd("go", "run", "scripts/embed/embed.go", "-pkg", "view", "-base", "server/view")
	embeds := map[string][]string{
		"css":       {"-var", "cssResources", "-out", "server/view/static-css.go", "-gzip", "css/*.css"},
		"images":    {"-var", "imageResources", "-out", "server/view/static-images.go", "images/*"},
		"js":        {"-var", "jsResources", "-out", "server/view/static-js.go", "-gzip", "js/*", "js/*/*"},
		"misc":      {"-var", "miscResources", "-out", "server/view/static-misc.go", "-gzip", "misc/*"},
		"templates": {"-var", "templateResources", "-out", "server/view/static-templates.go", "templates/*.html"},
	}

	for dir, args := range embeds {
		mod, err := target.Dir(args[3], filepath.Join("server", "view", dir))
		if err != nil {
			return err
		}
		if !mod {
			continue
		}
		if err := embed(args...); err != nil {
			return err
		}
	}
	return nil
}

// Install rebuilds everything and then go-installs.
func Install() error {
	mg.Deps(Embed, GenGoProtoStubs)
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
