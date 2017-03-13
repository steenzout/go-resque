package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/redis.v5"

	"github.com/steenzout/go-resque"
	"github.com/steenzout/go-resque/test/log"
)

const (
	// WorkerClass name of the worker struct.
	WorkerClass = "Multiplier"
)

// Multiplier Resque worker.
type Multiplier struct{}

// Perform multiplies the two given arguments.
func (p Multiplier) Perform(args ...resque.JobArgument) (interface{}, error) {
	x := args[0].(float64)
	y := args[1].(float64)

	return x * y, nil
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", server, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	log.Infof(Package, "Redis client set %v", client)

	worker, err := resque.NewWorker("Multiplier", client, &Multiplier{})
	if err != nil {
		panic(err)
	}
	log.Infof(Package, "queue %s created", worker.Queue.Name)
	log.Infof(Package, "worker %s created", worker.Name)

	var wg sync.WaitGroup
	chanIn := make(chan resque.Job, 1)
	chanErr := make(chan error, 1)
	chanQuit := make(chan bool, 1)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	chanOut := make(chan resque.Job, 1)
	chanErr2 := make(chan error, 1)
	chanQuit2 := make(chan bool, 1)

	chanOut2 := make(chan interface{}, 1)
	chanErr3 := make(chan error, 2)
	chanQuit3 := make(chan bool, 1)

	defer func() {
		wg.Wait()
		close(interrupt)
		close(chanIn)
		close(chanErr)
		close(chanQuit)
		close(chanOut)
		close(chanErr2)
		close(chanQuit2)
		close(chanOut2)
		close(chanErr3)
		close(chanQuit3)
	}()

	wg.Add(1)
	go worker.Produce(&wg, chanIn, chanErr, chanQuit)
	go worker.Consume(&wg, chanOut, chanErr2, chanQuit2)
	go worker.Process(&wg, chanOut2, chanErr3, chanQuit3)

	// wait time to wait between producing messages
	wait := time.Duration(2 * time.Second)

	// loop until we get an exit signal
	for {
		select {
		case killSignal := <-interrupt:
			// handle interrupt signal
			log.Infof(Package, "got signal %s", killSignal.String())
			chanQuit <- true
			chanQuit2 <- true
			chanQuit3 <- true
			return

		case err := <-chanErr:
			log.Errorf(Package, "Producer error: %s\n", err.Error())

		case err := <-chanErr2:
			log.Errorf(Package, "Consumer error: %s\n", err.Error())

		case err := <-chanErr3:
			log.Errorf(Package, "Perform error %s\n", err.Error())

		case job := <-chanOut:
			log.Infof(Package, "Message from queue = %s %v\n", job.Class, job.Args)

		case value := <-chanOut2:
			log.Infof(Package, "job output = %v", value)

		case <-time.After(wait):
			// queue new job
			args := make([]resque.JobArgument, 2)
			args[0] = 1
			args[1] = 2

			job := resque.Job{
				Class: WorkerClass,
				Args:  args,
			}
			chanIn <- job
			log.Infof(Package, "sent request to queue 1 job %s: %v", job.Class, job.Args)
		}
	}
}
