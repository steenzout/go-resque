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
	"encoding/json"
	"sync"

	"github.com/go-redis/redis"
)

// Worker primitive to implement a consumer.
type Worker struct {
	// JobClass type of job being of executed.
	JobClass JobClass
	// Done channel to request the worker to stop execution.
	Done <-chan bool
}

// Run worker until an error occurs or the worker is requested to quit.
func (w Worker) Run() error {
	pubsub, err := w.JobClass.Queue.Subscribe()
	if err != nil {
		return err
	}

	chan_in := make(chan Job, 1)
	chan_err := make(chan error, 1)
	chan_exit := make(chan bool, 1)

	var wg sync.WaitGroup

	defer func() {
		wg.Wait()
		close(chan_in)
		close(chan_err)
		close(chan_exit)
		pubsub.Close()
	}()

	go consume(pubsub, chan_in, chan_err, chan_exit, wg)
	chan_exit <- false

	for {
		select {
		case job := <-chan_in:
			w.JobClass.Performer.Perform(job.Args...)

			// ready for the next message
			chan_exit <- false

		case err = chan_err:
			chan_exit <- true
			return err

		case <-w.Done:
			chan_exit <- true
			return nil
		}
	}
}

// consume retrieves a message from a Redis subscription,
// transforms it into a Job and
// send to the input channel.
// In case of error during this process,
// it will send the error message to an error channel.
func consume(pubsub *redis.PubSub, chan_in <-chan Job, chan_err <-chan error, chan_exit <-chan bool, wg sync.WaitGroup) {
	defer wg.Done()

	var job Job

	for {
		select {
		case quit := <-chan_exit:
			if quit {
				return
			}

			for {
				msg, err := pubsub.ReceiveMessage()
				if err != nil {
					chan_err <- err
					continue
				}

				job = Job{}
				err = json.Unmarshal([]byte(msg.Payload), &job)
				if err != nil {
					chan_err <- err
					continue
				}

				chan_in <- job
			}
		}
	}
}
