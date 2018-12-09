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

import "syscall/js"

var (
	inputBroadcastOutputNum = doc.ElementByID("broadcast-outputnum")
	focusedBroadcast        *Broadcast
)

func init() {
	inputBroadcastOutputNum.AddEventListener("change", js.NewEventCallback(0, func(js.Value) {
		focusedBroadcast.OutputNum = uint(inputBroadcastOutputNum.Get("value").Int())
	}))
}

func (b *Broadcast) GainFocus() {
	focusedBroadcast = b
	inputBroadcastOutputNum.Set("value", b.OutputNum)
}
