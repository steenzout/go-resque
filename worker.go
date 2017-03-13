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

package resque

import (
	"sync"

	"gopkg.in/redis.v5"
)

// Worker primitive to implement producers or consumers.
type Worker struct {
	Name      string
	Queue     *Queue
	Performer Performer
}

// NewWorker creates a new JobClass pointer.
func NewWorker(n string, c *redis.Client, p Performer) *Worker {
	return &Worker{
		Name:      n,
		Queue:     newQueue(n, c),
		Performer: p,
	}
}

// Consume routine.
func (w Worker) Consume(wg sync.WaitGroup, chanOut chan Job, chanErr chan error, chanQuit <-chan bool) {
	defer wg.Done()

	for {
		select {
		case stop := <-chanQuit:
			if stop {
				return
			}

			job, err := w.Queue.Receive()
			if err != nil {
				chanErr <- err
			}

			chanOut <- *job
		}
	}
}

// Process routine.
func (w Worker) Process(wg *sync.WaitGroup, chanOut chan interface{}, chanErr chan error, chanQuit <-chan bool) {

	chan_in := make(chan Job, 1)
	chan_qerror := make(chan error, 1)
	chan_stop := make(chan bool, 1)

	wg.Add(1)
	defer func() {
		wg.Done()
		close(chan_in)
		close(chan_qerror)
		close(chan_stop)
	}()

	go w.Consume(wg, chan_in, chanErr, chan_stop)

	for {
		select {
		case <-chanQuit:
			chan_stop <- true
			return

		case err := <-chan_qerror:
			chanErr <- err

			// signal to get the next message
			chan_stop <- false

		case job := <-chan_in:
			out, err := w.Performer.Perform(job.Args...)
			if err != nil {
				chanErr <- err
			}

			chanOut <- out

			// signal to get the next message
			chan_stop <- false
		}
	}
}

// Produce routine.
func (w Worker) Produce(wg *sync.WaitGroup, chanIn <-chan Job, chanErr chan error, chanQuit <-chan bool) {

	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case quit := <-chanQuit:
			if quit {
				return
			}
		case job := <-chanIn:
			err := w.Queue.Send(job.Args)
			if err != nil {
				chanErr <- err
			}
		}
	}
}
