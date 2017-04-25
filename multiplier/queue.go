package multiplier

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
	"time"

	"gopkg.in/redis.v5"

	"github.com/steenzout/go-resque"
)

// Queue Multiplier queue.
type Queue struct {
	*resque.RedisQueue
}

// NewQueue creates a new multiplier queue.
func NewQueue(rc *redis.Client, timeout time.Duration, c *Consumer, p *Producer) (resque.Queue, error) {
	rq, err := resque.NewRedisQueue(
		Class,
		rc,
		timeout,
		c,
		p,
	)
	if err != nil {
		return nil, err
	}

	return &Queue{rq}, nil
}
