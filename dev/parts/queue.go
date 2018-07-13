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

package parts

import (
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

const queueTypeParam = "$Any"

var queuePins = pin.NewMap(
	&pin.Definition{
		Name:      "input",
		Direction: pin.Output,
		Type:      queueTypeParam,
	},
	&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      queueTypeParam,
	},
)

func init() {
	model.RegisterPartType("Queue", &model.PartType{
		New: func() model.Part { return &Queue{} },
		Panels: []model.PartPanel{
			{
				Name:   "Queue",
				Editor: `TODO(josh): Implement configuration UI for queues`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A Queue part reads and writes values as soon as possible.
				In the meantime, values are stored in a queue. The management
				of the queue is configurable.
			</p><p>
				Queues are either FIFO (first-in-first-out, or a traditional queue) 
				or LIFO (last-in-first-out, also known as a stack). Using a LIFO
				queue can have higher goodput than a FIFO queue.
			</p><p>
				Queues have a required size limit. The last-in item
				is dropped from the queue if reading an item puts the queue over
				the limit. A queue may temporarily use more memory than the limit.
			</p>
			</div>`,
			},
		},
	})
}

// QueueMode describes how to choose items to send.
type QueueMode string

// Valid values of QueueMode.
const (
	QueueModeFIFO QueueMode = "fifo"
	QueueModeLIFO QueueMode = "lifo"
)

// Queue is a basic queue part.
type Queue struct {
	Mode     QueueMode `json:"mode"`
	MaxItems int       `json:"max_items"`
}

// Clone returns a clone of this Queue.
func (q *Queue) Clone() model.Part {
	q0 := *q
	return &q0
}

func (m QueueMode) pick() string {
	switch m {
	case QueueModeFIFO:
		return "0"
	case QueueModeLIFO:
		return "len(queue)-1"
	default:
		panic("unknown mode " + m)
	}
}

func (m QueueMode) trim() string {
	switch m {
	case QueueModeFIFO:
		return "1:"
	case QueueModeLIFO:
		return ":idx"
	default:
		panic("unknown mode " + m)
	}
}

// Impl returns the Queue implementation.
func (q *Queue) Impl(types map[string]string) (head, body, tail string) {
	return fmt.Sprintf("const itemLim = %d", q.MaxItems),
		fmt.Sprintf(`
		queue := make([]%s, 0, itemLim)
		for {
			if len(queue) == 0 {
				if input == nil {
					break
				}
				queue = append(queue, <-input)
			}
			idx := %s
			out := queue[idx]
			select {
			case in, open := <-input:
				if !open {
					input = nil
					break // select
				}
				queue = append(queue, in)
				if len(queue) > itemLim {
					queue = queue[1:]
				}
			case output <- out:
				queue = queue[%s]
			}
		}`, types[queueTypeParam], q.Mode.pick(), q.Mode.trim()),
		"close(output)"
}

// Imports returns nil.
func (q *Queue) Imports() []string { return nil }

// Pins returns a map declaring a single output of any type.
func (q *Queue) Pins() pin.Map { return queuePins }

// TypeKey returns "Queue".
func (q *Queue) TypeKey() string { return "Queue" }
