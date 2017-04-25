package main

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
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/redis.v5"

	"github.com/steenzout/go-env"
	"github.com/steenzout/go-resque"
	"github.com/steenzout/go-resque/multiplier"
)

const (
	// Package package name.
	Package = "test.integration.main"
)

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", env.GetRedisHost(), env.GetRedisPort()),
		Password: env.GetRedisPassword(), // no password set
		DB:       0,                      // use default DB
	})
	defer client.Close()
	fmt.Printf("[%s] INFO Redis client set %v\n", Package, client)

	waitForMessage := time.Duration(10 * time.Second)

	consumer := multiplier.NewConsumer()
	producer := multiplier.NewProducer()
	queue, err := multiplier.NewQueue(client, waitForMessage, consumer, producer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[%s] INFO queue %s created\n", Package, queue.Name())

	pChanIn, pChanErr, pChanExit := getProducerChannels()
	cChanOut, cChanExit := getConsumerChannels()

	chanInterrupt := make(chan os.Signal, 1)
	signal.Notify(chanInterrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go producer.Publish(&wg, pChanIn, pChanErr, pChanExit)
	fmt.Printf("[%s] producer started...\n", Package)

	wg.Add(1)
	go consumer.Run(&wg, cChanOut, cChanExit)
	fmt.Printf("[%s] consumer started...\n", Package)

	defer func() {
		fmt.Println("waiting for routines")
		wg.Wait()
		close(pChanIn)
		close(pChanErr)
		close(pChanExit)
		close(chanInterrupt)
		close(cChanOut)
		close(cChanExit)
	}()

	// wait time to wait between producing messages
	wait := time.Duration(5 * time.Second)

	rand.Seed(42)

	// loop until we get an exit signal
	for {
		select {
		case killSignal := <-chanInterrupt:
			handleInterrupt(killSignal, pChanExit, cChanExit)
			return

		case err := <-pChanErr:
			fmt.Fprintf(os.Stderr, "[%s] ERROR Producer error: %s\n", Package, err.Error())

		case value := <-cChanOut:
			handleOutput(value, queue)

		case <-time.After(wait):
			sendJob(pChanIn, queue)
		}
	}
}

// getConsumerChannels initialize consumer channels.
func getConsumerChannels() (chan float64, chan bool) {
	cChanOut := make(chan float64, 1)
	cChanExit := make(chan bool, 1)
	return cChanOut, cChanExit
}

// getProducerChannels initialize producer channels.
func getProducerChannels() (chan *resque.Job, chan error, chan bool) {
	pChanIn := make(chan *resque.Job, 1)
	pChanErr := make(chan error, 1)
	pChanExit := make(chan bool, 1)
	return pChanIn, pChanErr, pChanExit
}

// handleOutput handle output from the consumer.
func handleOutput(output float64, queue resque.Queue) {
	fmt.Printf("[%s] INFO job output = %v\n", Package, output)

	size, err := queue.(*multiplier.Queue).Size()
	if err == nil {
		fmt.Printf("[%s] INFO queue %s has %d jobs\n", Package, queue.Name(), size)
	}
}

// handleInterrupt handle interrupt signals.
func handleInterrupt(killSignal os.Signal, pChanExit chan bool, cChanExit chan bool) {
	fmt.Printf("[%s] INFO main got signal %s\n", Package, killSignal.String())

	// broadcast exit signal to stop routines
	pChanExit <- true
	cChanExit <- true
}

// sendJob send a new Multiplier job with random arguments to the queue.
func sendJob(pChanIn chan *resque.Job, queue resque.Queue) {
	// generate a new Multiplier random job
	job := multiplier.NewJob(rand.Float64(), rand.Float64())

	// queue job
	pChanIn <- job
	fmt.Printf("[%s] INFO sent request to queue 1 job %s: %v\n", Package, job.Class, job.Args)

	// get queue size
	size, err := queue.(*multiplier.Queue).Size()
	if err != nil {
		fmt.Printf("[%s] ERROR %s\n", Package, err)
	}
	fmt.Printf("[%s] INFO queue %s has %d jobs\n", Package, queue.Name(), size)
}
