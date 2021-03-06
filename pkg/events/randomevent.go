// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import "time"

// RandomEvent a random event
type RandomEvent EventHappend

// Clone clones the Event
func (ce *RandomEvent) Clone() Event {
	clone := &RandomEvent{}
	clone.Etype = ce.Etype
	clone.Subject = ce.Subject
	clone.Time = ce.Time
	clone.Values = ce.Values
	return clone
}

// GetType returns type of an Event
func (ce *RandomEvent) GetType() EventType {
	return ce.Etype
}

// GetTime returns the time when the event occurs
func (ce *RandomEvent) GetTime() time.Time {
	return ce.Time
}

// GetValues returns the values of the event
func (ce *RandomEvent) GetValues() interface{} {
	return ce.Values
}

// GetSubject returns the subject of the event
func (ce *RandomEvent) GetSubject() string {
	return ce.Subject
}
