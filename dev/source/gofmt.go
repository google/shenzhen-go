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

package source

import (
	"go/format"
	"io"
	"io/ioutil"
)

// GoFmt reads all of src, applies "gofmt" to it, and then writes the result to dst.
func GoFmt(dst io.Writer, src io.Reader) error {
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
