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

// The embed script wraps file contents with some Go to access it.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	pkg  = flag.String("p", "", "Name of package, required")
	vn   = flag.String("v", "", "Name of map variable, required")
	outf = flag.String("o", "", "Name of output file, required")
)

func main() {
	flag.Parse()
	if *pkg == "" || *vn == "" || *outf == "" || len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	o, err := os.Create(*outf)
	if err != nil {
		log.Fatalf("Cannot create output file: %v", err)
	}
	defer o.Close()

	w := bufio.NewWriter(o)
	fmt.Fprintf(w, "// This file was generated by embed.go. DO NOT EDIT.\n\npackage %s\n\n", *pkg)
	fmt.Fprintf(w, "var %s = map[string][]byte{\n", *vn)

	for _, ifs := range flag.Args() {
		ms, err := filepath.Glob(ifs)
		if err != nil {
			log.Fatalf("Input pattern invalid: %v", err)
		}
		for _, m := range ms {
			i, err := ioutil.ReadFile(m)
			if err != nil {
				log.Printf("Cannot read input file: %v", err)
				continue
			}
			fmt.Fprintf(w, "\t%q: []byte(%q),\n", filepath.ToSlash(m), string(i))
		}
	}
	fmt.Fprintln(w, "}")

	if err := w.Flush(); err != nil {
		log.Fatalf("Cannot write output file: %v", err)
	}
	if err := o.Close(); err != nil {
		log.Fatalf("Cannot write output file: %v", err)
	}
}
