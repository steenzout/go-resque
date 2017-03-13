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
	"time"

	"gopkg.in/redis.v5"
)

// Worker primitive to implement producers or consumers.
type Worker struct {
	Name                string
	Queue               *Queue
	Performer           Performer
	waitBetweenMessages time.Duration
	waitForMessage      time.Duration
}

// NewWorker creates a new JobClass pointer.
func NewWorker(n string, c *redis.Client, p Performer) (*Worker, error) {
	q, err := newQueue(n, c)
	if err != nil {
		return nil, err
	}

	return &Worker{
		Name:                n,
		Queue:               q,
		Performer:           p,
		waitBetweenMessages: time.Duration(10 * time.Millisecond),
		waitForMessage:      time.Duration(1 * time.Second),
	}, nil
}

// Consume routine.
func (w Worker) Consume(wg *sync.WaitGroup, chanOut chan Job, chanErr chan error, chanQuit <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-chanQuit:
			return

		case <-time.After(w.waitBetweenMessages):

			job, err := w.Queue.Receive()
			if err != nil {
				chanErr <- err
			}

			if job == nil {
				time.Sleep(w.waitForMessage)
			} else {
				chanOut <- *job
			}
		}
	}
}

// Process routine.
func (w Worker) Process(wg *sync.WaitGroup, chanOut chan interface{}, chanErr chan error, chanQuit <-chan bool) {

	chanIn := make(chan Job, 1)
	chanQErr := make(chan error, 1)
	chanStop := make(chan bool, 1)

	defer func() {
		wg.Done()
		close(chanIn)
		close(chanQErr)
		close(chanStop)
	}()

	go w.Consume(wg, chanIn, chanErr, chanStop)

	for {
		select {
		case <-chanQuit:
			chanStop <- true
			return

		case err := <-chanQErr:
			chanErr <- err

			// signal to get the next message
			chanStop <- false

		case job := <-chanIn:
			out, err := w.Performer.Perform(job.Args...)
			if err != nil {
				chanErr <- err
			}

			chanOut <- out

			// signal to get the next message
			chanStop <- false
		}
	}
}

// Produce routine.
func (w Worker) Produce(wg *sync.WaitGroup, chanIn <-chan Job, chanErr chan error, chanQuit <-chan bool) {

	defer wg.Done()

	for {
		select {
		case <-chanQuit:
			return
		case job := <-chanIn:
			err := w.Queue.Send(job.Args)
			if err != nil {
				chanErr <- err
			}
		}
	}
}
