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
	"encoding/json"
	"fmt"

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
}

// newQueue initializes a Queue struct and updates the set of available Resque queues.
func newQueue(jcn string, c *redis.Client) (*Queue, error) {
	q := &Queue{
		redis:        c,
		jobClassName: jcn,
		Name:         fmt.Sprintf("resque:queue:%s", jcn),
	}

	exists, err := c.SIsMember(Queues, q.Name).Result()
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err := c.SAdd(Queues, q.Name).Result()
		if err != nil {
			return nil, err
		}
	}

	return q, nil
}

// Receive gets a job from the queue.
func (q Queue) Receive() (*Job, error) {
	cmd := q.redis.LPop(q.Name)
	if cmd.Err() != nil {
		if cmd.Err().Error() == "redis: nil" {
			return nil, nil
		}
		return nil, cmd.Err()
	}

	jsonStr, err := cmd.Bytes()
	if err != nil {
		return nil, err
	}

	job := &Job{}
	err = json.Unmarshal(jsonStr, job)
	if err != nil {
		return nil, err
	}

	return job, err
}

// Send places a job on the queue.
func (q Queue) Send(args []JobArgument) error {
	jsonStr, err := json.Marshal(Job{Class: q.jobClassName, Args: args})
	if err != nil {
		return err
	}

	return q.redis.RPush(q.Name, jsonStr).Err()
}
