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
	"github.com/steenzout/go-resque"
)

const (
	// Class Resque class name.
	Class = "Multiplier"
)

// Job representation of a multiplier job message.
type Job struct {
	// Arg1 first argument.
	Arg1 float64
	// Arg2 second argument.
	Arg2 float64
}

// NewJob returns a new multiplier job message.
func NewJob(arg1, arg2 float64) *resque.Job {
	return &resque.Job{
		Class: Class,
		Args: []resque.JobArgument{
			arg1,
			arg2,
		},
	}
}

// NewJobFromArgs builds the representation of the multiplier job message.
func NewJobFromArgs(args ...resque.JobArgument) *Job {
	return &Job{
		Arg1: args[0].(float64),
		Arg2: args[1].(float64),
	}
}
