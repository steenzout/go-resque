package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gopkg.in/redis.v5"

	"github.com/steenzout/go-resque"
	"github.com/steenzout/go-resque/test/log"
)

// Multiplier Resque worker.
type Multiplier struct{}

// Perform multiplies the two given arguments.
func (p Multiplier) Perform(args ...resque.JobArgument) (interface{}, error) {
	x := args[0].(int)
	y := args[1].(int)

	return x * y, nil
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", server, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	worker := resque.NewWorker("Multiplier", client, &Multiplier{})

	var wg sync.WaitGroup
	chanIn := make(chan resque.Job, 1)
	chanErr := make(chan error, 1)
	chanQuit := make(chan bool, 1)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	defer func() {
		wg.Wait()
		close(interrupt)
		close(chanIn)
		close(chanErr)
		close(chanQuit)
	}()

	wg.Add(1)
	go worker.Produce(&wg, chanIn, chanErr, chanQuit)

	// loop until we get an exit signal
	for {
		select {
		case killSignal := <-interrupt:
			log.Infof(Package, "got signal %s", killSignal.String())
			chanQuit <- true
			return
		}
	}
}
