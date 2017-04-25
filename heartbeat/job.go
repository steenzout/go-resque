package heartbeat

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

	"github.com/steenzout/go-resque"
)

const (
	// Class Resque class name.
	Class = "Heartbeat"
)

// Job representation of a hearbeat job message.
type Job struct {
	// ID daemon process unique identifier.
	ID string
	// name name of the daemon.
	Name string
	// Timestamp message timestamp, in RFC3339 format.
	Timestamp string
}

// NewJob returns a new heartbeat job message.
func NewJob(daemon, id string) *resque.Job {
	return &resque.Job{
		Class: Class,
		Args: []resque.JobArgument{
			id,
			daemon,
			time.Now().Format(time.RFC3339),
		},
	}
}

// NewJobFromArgs builds the representation of the heartbeat job message.
func NewJobFromArgs(args ...resque.JobArgument) *Job {
	return &Job{
		ID:        args[0].(string),
		Name:      args[1].(string),
		Timestamp: args[2].(string),
	}
}
