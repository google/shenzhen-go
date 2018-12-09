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

package dom

import (
	"fmt"
	"strconv"
	"syscall/js"
)

// Float converts a js.Value to a float. If the Value is a Float, it calls x.Float.
// If the value is a string, it uses ParseFloat. Otherwise, it returns an error.
func Float(x js.Value) (float64, error) {
	switch x.Type() {
	case js.TypeNumber:
		return x.Float(), nil
	case js.TypeString:
		return strconv.ParseFloat(x.String(), 64)
	default:
		return 0, fmt.Errorf("value not convertible to float: %v", x)
	}
}
