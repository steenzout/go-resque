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

import (
	"sync"
)

// Producer interface Resque job producers must implement.
type Producer interface {
	// Produce creates a Resque job.
	Produce(args ...JobArgument) *Job
	// Queue returns the queue.
	Queue() Queue
	// SetQueue assign a queue.
	SetQueue(q Queue)
}

// BaseProducer Base struct for a job producer.
type BaseProducer struct {
	*Worker
}

// NewProducer creates a new producer pointer.
func NewProducer() *BaseProducer {
	return &BaseProducer{
		NewWorker(),
	}
}

// Produce creates a Resque job message.
func (bp *BaseProducer) Produce(args ...JobArgument) *Job {
	return &Job{
		Class: bp.queue.Name(),
		Args:  args,
	}
}

// Publish Go routine that sends Resque jobs to the worker Queue.
func (bp *BaseProducer) Publish(wg *sync.WaitGroup, chanIn <-chan *Job, chanErr chan error, chanExit <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-chanExit:
			return
		case job := <-chanIn:
			err := bp.queue.Send(job)
			if err != nil {
				chanErr <- err
			}
		}
	}
}
