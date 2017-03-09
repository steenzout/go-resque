package resque

import (
	"encoding/json"
	"sync"

	"github.com/go-redis/redis"
)

// Performer interface that each consumer/worker must implement.
type Performer interface {
	// Perform consumes the given payload.
	Perform(args ...JobArgument)
}

// Worker primitive to implement a consumer.
type Worker struct {
	JobClass JobClass
	Done     <-chan bool
}

// Run .
func (w Worker) Run() error {
	var wg sync.WaitGroup

	pubsub, err := w.JobClass.Queue.Subscribe()
	if err != nil {
		return err
	}
	input := make(chan Job, 1)

	defer func() {
		wg.Wait()
		pubsub.Close()
	}()

	go consume(pubsub, input)

	for {
		select {
		case job := <-input:
			w.JobClass.Performer.Perform(job.Args...)
		case <-w.Done:
			close(input)
			return nil
		}
	}
}

func consume(pubsub *redis.PubSub, c <-chan Job) {
	var job Job

	defer func() { recover() }()

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			continue
		}

		job = Job{}
		err = json.Unmarshal([]byte(msg.Payload), &job)
		if err != nil {
			continue
		}

		c <- job
	}

	return
}
