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
	"os"
	"os/user"
	"path/filepath"
)

// GoPath returns the GOPATH using Go 1.8 and later rules.
// If the GOPATH environment var is defined, it uses that, otherwise it
// assumes $HOME/go.
func GoPath() (string, error) {
	// TODO: Any implementation of this in the std lib?
	p, _ := os.LookupEnv("GOPATH")
	if p != "" {
		return p, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, "go"), nil
}
