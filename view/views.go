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

// Package view provides the user interface.
package view

import (
	"io"
	"os/exec"
)

const css = `
	body {
		font-family: "Go","San Francisco","Helvetica Neue",Helvetica,sans-serif;
		float: none;
		max-width: 800px;
		margin: 20 auto 0;
	}
	form {
		float: none;
		max-width: 800px;
		margin: 0 auto;
	}
	div.formfield {
		margin-top: 12px;
		margin-bottom: 12px;
	}
	label {
		float: left;
		text-align: right;
		margin-right: 15px;
		width: 30%;
	}
	input {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	input[type=text] {
		width: 65%;
	}
	select {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	textarea {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	div svg {
		display: block;
		margin: 0 auto;
	}
	div.hcentre {
		text-align: center;
	}
	table.browse {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
		margin-top: 16pt;
	}
`

// TODO: Deduplicate.
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
