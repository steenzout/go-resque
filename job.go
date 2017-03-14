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
)

// JobArgument represent a single argument.
type JobArgument interface{}

// Job represents a unit of work.
// Each job lives on a single queue and has an associated payload object.
//
// The payload is a hash with two attributes: `class` and `args`.
// The `class` is the name of the class which
// should be used to run the job.
//
// `Args` are an array of arguments which
// should be passed to the class's `perform` function.
//
type Job struct {
	Class string        `json:"class"`
	Args  []JobArgument `json:"args"`
}

// MarshalBinary encodes itself into a binary format.
func (j Job) MarshalBinary() ([]byte, error) {
	return json.Marshal(j)
}

// UnmarshalBinary decodes the binary representation into itself.
func (j *Job) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, j)
}
