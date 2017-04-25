package resque

import (
	"sync"
)

// Consumer interface Resque job consumers must implement.
type Consumer interface {
	// Consume performs the Resque job work.
	Consume(args ...JobArgument) (interface{}, error)
	// Queue returns the queue.
	Queue() Queue
	// SetQueue assign a queue.
	SetQueue(q Queue)
}

// BaseConsumer Base struct for a Resque job consumer.
type BaseConsumer struct {
	*Worker
}

// NewConsumer creates a new consumer pointer.
func NewConsumer() *BaseConsumer {
	return &BaseConsumer{
		NewWorker(),
	}
}

// Consume pass job class and arguments to the output.
func (bc *BaseConsumer) Consume(args ...JobArgument) (interface{}, error) {
	return &Job{
		Class: bc.queue.Name(),
		Args:  args,
	}, nil
}

// Subscribe Go routine that retrieves Resque jobs from the worker Queue.
func (bc *BaseConsumer) Subscribe(wg *sync.WaitGroup, chanOut chan *Job, chanErr chan error, chanNext <-chan bool, chanExit <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-chanExit:
			return

		case <-chanNext:
			job, err := bc.queue.Receive()
			if err != nil {
				chanErr <- err
				continue
			}

			chanOut <- job
		}
	}
}
