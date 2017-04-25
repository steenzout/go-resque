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
	qRedis := queue.(*multiplier.Queue)
	fmt.Printf("[%s] INFO queue %s created\n", Package, queue.Name())

	pChanIn := make(chan *resque.Job, 1)
	pChanErr := make(chan error, 1)
	pChanExit := make(chan bool, 1)

	cChanOut := make(chan float64, 1)
	cChanExit := make(chan bool, 1)

	chanInterrupt := make(chan os.Signal, 1)
	signal.Notify(chanInterrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go producer.Publish(&wg, pChanIn, pChanErr, pChanExit)
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
	fmt.Printf("[%s] producer started...\n", Package)

	wg.Add(1)
	go consumer.Run(&wg, cChanOut, cChanExit)
	fmt.Printf("[%s] consumer started...\n", Package)

	// wait time to wait between producing messages
	wait := time.Duration(5 * time.Second)

	rand.Seed(42)

	// loop until we get an exit signal
	for {
		select {
		case killSignal := <-chanInterrupt:
			// handle chanInterrupt signal
			fmt.Printf("[%s] INFO main got signal %s\n", Package, killSignal.String())
			pChanExit <- true
			cChanExit <- true
			return

		case err := <-pChanErr:
			fmt.Fprintf(os.Stderr, "[%s] ERROR Producer error: %s\n", Package, err.Error())

		case value := <-cChanOut:
			fmt.Printf("[%s] INFO job output = %v\n", Package, value)

			size, err := qRedis.Size()
			if err == nil {
				fmt.Printf("[%s] INFO queue %s has %d jobs\n", Package, queue.Name(), size)
			}

		case <-time.After(wait):
			// generate a new Multiplier random job
			job := multiplier.NewJob(rand.Float64(), rand.Float64())

			// queue job
			pChanIn <- job

			fmt.Printf("[%s] INFO sent request to queue 1 job %s: %v\n", Package, job.Class, job.Args)
			size, err := qRedis.Size()
			if err != nil {
				fmt.Printf("[%s] ERROR %s\n", Package, err)
			}
			fmt.Printf("[%s] INFO queue %s has %d jobs\n", Package, queue.Name(), size)
		}
	}
}
