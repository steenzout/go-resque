package resque

//
// Copyright 2017 Pedro Salgado
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Worker base struct to build consumers and producers.
type Worker struct {
	queue Queue
}

// NewWorker creates a new worker pointer.
func NewWorker() *Worker {
	return &Worker{}
}

// Queue returns the queue.
func (w *Worker) Queue() Queue {
	return w.queue
}

// SetQueue assign a Queue.
func (w *Worker) SetQueue(q Queue) {
	w.queue = q
}
