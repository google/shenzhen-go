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

package graph

import (
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func pipeThru(dst io.Writer, cmd *exec.Cmd, src io.Reader) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go io.Copy(os.Stderr, stderr)
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
	in, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}
	out, err := format.Source(in)
	if err != nil {
		return err
	}
	_, err = dst.Write(out)
	return err
}

func goimports(dst io.Writer, src io.Reader) error {
	// TODO: Use golang.org/x/tools/imports package instead.
	return pipeThru(dst, exec.Command(`goimports`), src)
}
