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

//+build js

package parts

import "github.com/google/shenzhen-go/dom"

var (
	inputZipInputNum    = doc.ElementByID("zip-inputnum")
	selectZipFinishMode = doc.ElementByID("zip-finishmode")
	focusedZip          *Zip
)

func init() {
	inputZipInputNum.AddEventListener("change", dom.NewEventCallback(0, func(dom.Object) {
		focusedZip.InputNum = uint(inputZipInputNum.Get("value").Int())
	}))
	selectZipFinishMode.AddEventListener("change", dom.NewEventCallback(0, func(dom.Object) {
		focusedZip.FinishMode = ZipFinishMode(selectZipFinishMode.Get("value").String())
	}))
}

func (z *Zip) GainFocus() {
	focusedZip = z
	inputZipInputNum.Set("value", z.InputNum)
	selectZipFinishMode.Set("value", z.FinishMode)
}
