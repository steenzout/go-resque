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

package resque

import (
	"fmt"
	"time"

	"gopkg.in/redis.v5"
)

const (
	// Queues name of the key for the set where Resque stores the currently available queues.
	Queues = "resque:queues"
)

// Queue a job queue.
type Queue struct {
	redis        *redis.Client
	jobClassName string
	Name         string
	timeout      time.Duration
}

// NewQueue initializes a Queue struct and updates the set of available Resque queues.
func NewQueue(jcn string, c *redis.Client, timeout time.Duration) (*Queue, error) {
	q := &Queue{
		redis:        c,
		jobClassName: jcn,
		Name:         fmt.Sprintf("resque:queue:%s", jcn),
		timeout:      timeout,
	}

	exists, err := c.SIsMember(Queues, jcn).Result()
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err := c.SAdd(Queues, jcn).Result()
		if err != nil {
			return nil, err
		}
	}

	return q, nil
}

// Peek returns from the queue the jobs at position(s) [start, stop].
func (q Queue) Peek(start, stop int64) ([]Job, error) {

	var jobs []Job

	err := q.redis.LRange(q.Name, start, stop).ScanSlice(jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// Receive gets a job from the queue.
func (q Queue) Receive() (*Job, error) {

	var job Job

	v, err := q.redis.BLPop(q.timeout, q.Name).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}

	err = job.UnmarshalBinary([]byte(v[1]))
	if err != nil {
		return nil, err
	}

	return &job, err
}

// Send places a job to the queue.
func (q Queue) Send(job Job) error {

	byteArr, err := job.MarshalBinary()
	if err != nil {
		return err
	}

	return q.redis.RPush(q.Name, string(byteArr)).Err()
}

// Size returns the number of jobs in the queue.
func (q Queue) Size() (int64, error) {
	v, err := q.redis.LLen(q.Name).Result()
	if err != nil {
		return 0, err
	}

	return v, err
}
