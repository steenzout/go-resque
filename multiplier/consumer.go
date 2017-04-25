package multiplier

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
	"fmt"
	"os"
	"sync"

	"github.com/steenzout/go-resque"
)

// Consumer Multiplier message consumer.
type Consumer struct {
	*resque.BaseConsumer
}

// NewConsumer returns a new Consumer pointer.
func NewConsumer() *Consumer {
	return &Consumer{
		resque.NewConsumer(),
	}
}

// Consume returns the multiplication of the given arguments.
func (c *Consumer) Consume(args ...resque.JobArgument) (interface{}, error) {
	fmt.Printf("[multiplier/consumer] DEBUG args %v\n", args)

	job := NewJobFromArgs(args...)
	output := job.Arg1 * job.Arg2

	fmt.Printf("[multiplier/consumer] DEBUG %f * %f = %f\n", job.Arg1, job.Arg2, output)

	return output, nil
}

// Run routine that prints the results from the Multiplier queue.
func (c *Consumer) Run(wg *sync.WaitGroup, chanOut chan float64, chanExit <-chan bool) {
	defer wg.Done()

	chanJob, chanErr, chanQuit, chanNext := getChannels()

	var localWG sync.WaitGroup
	localWG.Add(1)
	go c.Subscribe(&localWG, chanJob, chanErr, chanNext, chanQuit)
	defer func() {
		fmt.Println("[multiplier/consumer.go] INFO waiting for subscription to timeout")
		localWG.Wait()
		close(chanNext)
		close(chanJob)
		close(chanErr)
		close(chanQuit)
	}()

	// signal ready to receive messages
	chanNext <- true

	for {
		select {
		case <-chanExit:
			chanQuit <- true
			fmt.Println("multiplier/consumer.go: exit")
			return

		case err := <-chanErr:
			fmt.Fprintf(os.Stderr, "[multiplier/consumer] ERROR %s\n", err.Error())
			chanNext <- true

		case job := <-chanJob:
			fmt.Printf("[multiplier/consumer] got job %s\n", job)
			out, err := c.Consume(job.Args...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[multiplier] ERROR %s\n", err.Error())
				chanNext <- true
				continue
			}
			fmt.Printf("[multiplier/consumer] INFO %f\n", out)
			chanOut <- out.(float64)
			chanNext <- true
		}
	}
}

// getChannels creates channels to communicate with the routine.
func getChannels() (chan *resque.Job, chan error, chan bool, chan bool) {
	chanJob := make(chan *resque.Job, 1)
	chanErr := make(chan error, 1)
	chanQuit := make(chan bool, 1)
	// channel to signal this go routine is ready to process the next message
	chanNext := make(chan bool, 1)
	return chanJob, chanErr, chanQuit, chanNext
}
