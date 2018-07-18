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
		Direction: pin.Input,
		Type:      queueTypeParam,
	},
	&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      queueTypeParam,
	},
	&pin.Definition{
		Name:      "drop",
		Direction: pin.Output,
		Type:      queueTypeParam,
	},
)

func init() {
	model.RegisterPartType("Queue", "Flow", &model.PartType{
		New: func() model.Part {
			return &Queue{
				Mode:     QueueModeLIFO,
				MaxItems: 1000,
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Queue",
				Editor: `
			<div class="form">
				<div class="formfield">
					<label for="queue-maxitems">Max items</label>
					<input id="queue-maxitems" name="queue-maxitems" type="number" required title="Must be a whole number, at least 1." value="1"></input>
				</div>
				<div class="formfield">
					<label for="queue-mode">Mode</label>
					<select id="queue-mode" name="queue-mode">
						<option value="lifo" selected>LIFO (stack)</option>
						<option value="fifo">FIFO (queue)</option>
					</select>
				</div>
			</div>`,
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
				or LIFO (last-in-first-out, also known as a stack).
			</p><p>
				Using a LIFO
				queue can have higher goodput than a FIFO queue.
			</p><p>
				Queues have a required maximum number of items. If reading an item 
				puts the queue over	the limit, the least recently read item is dropped
				from the queue, rather than waiting for the queue to lower. 
				Dropped items are sent to the drop output, but unlike the main output,
				the queue will not block on sending to drop.
				A queue may temporarily use more memory than the limit.
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

func (m QueueMode) params() (index, trim string) {
	switch m {
	case QueueModeFIFO:
		return "0", "1:"
	case QueueModeLIFO:
		return "len(queue)-1", ":idx"
	default:
		panic("unknown mode " + m)
	}
}

// Impl returns the Queue implementation.
func (q *Queue) Impl(n *model.Node) model.PartImpl {
	index, trim := q.Mode.params()
	return model.PartImpl{
		Head: fmt.Sprintf("const maxItems = %d", q.MaxItems),
		Body: fmt.Sprintf(`
		queue := make([]%s, 0, maxItems)
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
				if len(queue) <= maxItems {
					break // select
				}
				// Drop least-recently read item, but don't block.
				select {
				case drop <- queue[0]:
				default:
				}
				queue = queue[1:]
			case output <- out:
				queue = queue[%s]
			}
		}`, n.TypeParams[queueTypeParam], index, trim),
		Tail: `close(output)
		if drop != nil {
			close(drop)
		}`,
	}
}

// Pins returns a map declaring an input and two outputs of the same arbitrary type.
func (q *Queue) Pins() pin.Map { return queuePins }

// TypeKey returns "Queue".
func (q *Queue) TypeKey() string { return "Queue" }
