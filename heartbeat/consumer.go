package heartbeat

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
	"time"

	"github.com/steenzout/go-resque"
)

// Consumer Hearbeat message consumer.
type Consumer struct {
	*resque.BaseConsumer
}

// NewConsumer returns a new Consumer pointer.
func NewConsumer() *Consumer {
	return &Consumer{
		resque.NewConsumer(),
	}
}

// Consume perform the Resque job work.
func (c *Consumer) Consume(args ...resque.JobArgument) (interface{}, error) {
	job := NewJobFromArgs(args)

	rfc3339Ts, err := time.Parse(time.RFC3339, job.Timestamp)
	if err != nil {
		fmt.Fprintf(os.Stdout, "[heartbeat] INFO daemon %s id %s timestamp %s\n", job.Name, job.ID, job.Timestamp)
		return nil, fmt.Errorf("timestamp %s is not in the RFC3339 format", job.Timestamp)
	}

	fmt.Fprintf(os.Stdout, "[heartbeat] INFO daemon %s id %s timestamp %s\n", job.Name, job.ID, rfc3339Ts)

	return job, nil
}

// Run routine that prints messages from the Heartbeat queue.
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
	defer localWG.Wait()

	localWG.Add(1)
	go c.Subscribe(&localWG, chanRJob, chanRErr, chanRNext, chanRExit)

	// signal ready to receive messages
	chanRNext <- true

	for {
		select {
		case <-chanExit:
			chanRExit <- true
			return

		case err := <-chanRErr:
			fmt.Fprintf(os.Stderr, "[heartbeat] ERROR %s\n", err.Error())
			chanRNext <- true

		case job := <-chanRJob:
			_, err := c.Consume(job.Args...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[heartbeat] ERROR %s\n", err.Error())
				chanRNext <- true
				continue
			}
			chanRNext <- true
		}
	}
}
