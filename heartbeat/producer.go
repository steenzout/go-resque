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
	"sync"
	"time"

	"github.com/steenzout/go-resque"
)

// Producer Hearbeat message producer.
type Producer struct {
	*resque.BaseProducer
	// ID daemon process unique identifier.
	ID string
	// name name of the daemon.
	Name string
	// Interval time between heartbeat messages.
	Interval time.Duration
}

// NewProducer returns a new Producer pointer.
func NewProducer(id, name string, interval time.Duration) *Producer {
	return &Producer{
		BaseProducer: resque.NewProducer(),
		ID:           id,
		Name:         name,
		Interval:     interval,
	}
}

// Run routine that sends heartbeat messages to the worker's queue at the configured regular interval.
func (w *Producer) Run(wg *sync.WaitGroup, chanErr chan error, chanExit <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-chanExit:
			return

		case <-time.After(w.Interval):

			err := w.Queue().Send(NewJob(w.Name, w.ID))
			if err != nil {
				chanErr <- err
			}
		}
	}
}
