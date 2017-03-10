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

	"github.com/go-redis/redis"
)

// Queue a job queue.
type Queue struct {
	Name  string
	Redis *redis.Client
}

// NewQueue links a job queue to a Redis instance.
func NewQueue(n string, c *redis.Client) *Queue {
	return &Queue{
		Name:  n,
		Redis: c,
	}
}

// Put places a job on the queue.
func (q *Queue) Put(j *Job) error {
	json, err := json.Marshal(j)
	if err != nil {
		return err
	}

	cmd := q.Redis.RPush(q.Name, json)
	return cmd.Err()
}

// Subscribe returns a queue subscriber.
func (q *Queue) Subscribe() (*redis.PubSub, error) {
	pubsub, err := q.Redis.Subscribe(q.Name)
	if err != nil {
		return nil, err
	}

	return pubsub, nil
}
