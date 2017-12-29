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

package jsutil

const svgNamespaceURI = "http://www.w3.org/2000/svg"

// MakeTextNode creates a text node element in the global document.
func MakeTextNode(text string) Element {
	return Wrap(Document.Call("createTextNode", text))
}

// MakeSVGElement creates an element in the global document,
// belonging to the the SVG NS (http://www.w3.org/2000/svg).
func MakeSVGElement(n string) Element {
	return Wrap(Document.Call("createElementNS", svgNamespaceURI, n))
}
