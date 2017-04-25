package resque

import "sync"

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
			if chanErr != nil {
				chanErr <- err
			}
		}
	}
}
