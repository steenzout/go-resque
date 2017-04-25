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

// Queue Queue interface.
type Queue interface {
	Consumer() Consumer
	Name() string
	Producer() Producer
	Receive() (*Job, error)
	Send(job *Job) error
}

// RedisQueue a job RedisQueue.
type RedisQueue struct {
	redis   *redis.Client
	name    string
	timeout time.Duration
	// queueName name of the Queue.
	queueName string
	// Consumer Resque job consumer.
	consumer Consumer
	// Producer Resque job producer.
	producer Producer
}

// NewRedisQueue initializes a RedisQueue struct and updates the set of available Resque queues.
func NewRedisQueue(name string, rc *redis.Client, timeout time.Duration, c Consumer, p Producer) (*RedisQueue, error) {
	q := &RedisQueue{
		redis:     rc,
		name:      name,
		queueName: fmt.Sprintf("resque:Queue:%s", name),
		timeout:   timeout,
		consumer:  c,
		producer:  p,
	}
	if c != nil {
		c.SetQueue(q)
	}
	if p != nil {
		p.SetQueue(q)
	}

	exists, err := rc.SIsMember(Queues, name).Result()
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err := rc.SAdd(Queues, name).Result()
		if err != nil {
			return nil, err
		}
	}

	return q, nil
}

// Consumer returns this Queue's consumer.
func (rq *RedisQueue) Consumer() Consumer {
	return rq.consumer
}

// Name returns the Queue name.
func (rq *RedisQueue) Name() string {
	return rq.name
}

// Producer returns this Queue's producer.
func (rq *RedisQueue) Producer() Producer {
	return rq.producer
}

// Peek returns from the RedisQueue the jobs at position(s) [start, stop].
func (rq *RedisQueue) Peek(start, stop int64) ([]Job, error) {

	var jobs []Job

	err := rq.redis.LRange(rq.queueName, start, stop).ScanSlice(jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// Receive gets a job from the RedisQueue.
func (rq *RedisQueue) Receive() (*Job, error) {

	var job Job

	v, err := rq.redis.BLPop(rq.timeout, rq.queueName).Result()
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

// Send places a job to the RedisQueue.
func (rq *RedisQueue) Send(job *Job) error {

	byteArr, err := job.MarshalBinary()
	if err != nil {
		return err
	}

	return rq.redis.RPush(rq.queueName, string(byteArr)).Err()
}

// Size returns the number of jobs in the RedisQueue.
func (rq *RedisQueue) Size() (int64, error) {
	v, err := rq.redis.LLen(rq.queueName).Result()
	if err != nil {
		return 0, err
	}

	return v, err
}
