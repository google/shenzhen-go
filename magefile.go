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
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

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
	return sh.RunWith(env, "gopherjs", "build", "-o", "server/view/js/client.js", "github.com/google/shenzhen-go/client")
}

// Embeds static content into static-*.go files in the server/view package.
func Embed() error {
	mg.Deps(BuildClient)

	embed := sh.RunCmd("go", "run", "scripts/embed/embed.go")

	if err := embed("-p", "view", "-v", "cssResources", "-o", "server/view/static-css.go", "-base", "server/view", "css/*.css"); err != nil {
		return err
	}
	if err := embed("-p", "view", "-v", "imageResources", "-o", "server/view/static-images.go", "-base", "server/view", "images/*"); err != nil {
		return err
	}
	if err := embed("-p", "view", "-v", "jsResources", "-o", "server/view/static-js.go", "-base", "server/view", "js/*", "js/*/*"); err != nil {
		return err
	}
	if err := embed("-p", "view", "-v", "miscResources", "-o", "server/view/static-misc.go", "-base", "server/view", "misc/*"); err != nil {
		return err
	}
	return embed("-p", "view", "-v", "templateResources", "-o", "server/view/static-templates.go", "-base", "server/view", "templates/*.html")
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
