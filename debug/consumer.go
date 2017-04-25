package debug

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

// Consumer Debug message consumer.
type Consumer struct {
	*resque.BaseConsumer
}

// Consume prints the contents of the Resque job message.
func (c *Consumer) Consume(args ...resque.JobArgument) (interface{}, error) {
	fmt.Printf("[debug] DEBUG queue %s arguments %v", c.Queue().Name(), args)
	return nil, nil
}

// NewConsumer returns a new Consumer pointer.
func NewConsumer() *Consumer {
	return &Consumer{
		resque.NewConsumer(),
	}
}

// Run routine that prints messages from a Resque queue.
func (c *Consumer) Run(wg *sync.WaitGroup, chanExit <-chan bool) {
	defer wg.Done()

	chanRJob := make(chan *resque.Job, 1)
	defer close(chanRJob)

	chanRErr := make(chan error, 1)
	defer close(chanRErr)

	chanRExit := make(chan bool, 1)
	defer close(chanRExit)

	// channel to signal this go routine is ready to process the next message
	chanRNext := make(chan bool, 1)
	defer close(chanRNext)

	var localWG sync.WaitGroup

	localWG.Add(1)
	go c.Subscribe(&localWG, chanRJob, chanRErr, chanRNext, chanRExit)
	defer localWG.Wait()

	// signal ready to receive messages
	chanRNext <- true

	for {
		select {
		case <-chanExit:
			chanRExit <- true
			return

		case err := <-chanRErr:
			fmt.Fprintf(os.Stderr, "[debug] ERROR %s\n", err.Error())
			chanRNext <- true

		case job := <-chanRJob:
			_, err := c.Consume(job.Args...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[debug] ERROR %s\n", err.Error())
				chanRNext <- true
				continue
			}
			chanRNext <- true
		}
	}
}
