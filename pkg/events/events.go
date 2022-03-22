// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package events define a generic type of event for implementing of
// an event distribution mechanism.
package events

import (
	"time"
)

// EventType is an enumeration of the kind of events that can occur.
type EventType uint16

// Values of the EventType enumeration
const ( // For event types
	EventTypeConfiguration EventType = iota
	EventTypeOperationalState
	EventTypeRandom
)

func (et EventType) String() string {
	return [...]string{"Configuration", "OperationalState", "RandomEvent"}[et]
}

// Event an interface which defines the Event methods
type Event interface {
	GetType() EventType
	GetTime() time.Time
	GetValues() interface{}
	GetSubject() string
	Clone() Event
}

// EventHappend is a general purpose base type of event
type EventHappend struct {
	Subject string
	Time    time.Time
	Etype   EventType
	Values  interface{}
	Client  interface{}
}

// Clone clones the Event
func (eh *EventHappend) Clone() Event {
	clone := &EventHappend{}
	clone.Etype = eh.Etype
	clone.Subject = eh.Subject
	clone.Time = eh.Time
	clone.Values = eh.Values
	return clone
}

// GetType returns type of an Event
func (eh *EventHappend) GetType() EventType {
	return eh.Etype
}

// GetTime returns the time when the event occurs
func (eh *EventHappend) GetTime() time.Time {
	return eh.Time
}

// GetValues returns the values of the event
func (eh *EventHappend) GetValues() interface{} {
	return eh.Values
}

// GetSubject returns the subject of the event
func (eh *EventHappend) GetSubject() string {
	return eh.Subject
}
