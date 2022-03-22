// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
