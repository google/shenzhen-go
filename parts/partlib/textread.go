// Copyright 2017 Google Inc.
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

package partlib

import (
	"bufio"
	"os"
)

// StreamTextFile tries to read the file at the given path, and streams
// text lines from the file as string to the output.
func StreamTextFile(path string, output chan<- string, errors chan<- error) {
	f, err := os.Open(path)
	if err != nil {
		errors <- err
		return
	}
	defer f.Close()
	// TODO: switch to bufio.Reader since it can handle longer lines.
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		output <- sc.Text()
	}
	if err := sc.Err(); err != nil {
		errors <- err
		return
	}
	if err := f.Close(); err != nil {
		errors <- err
	}
}
