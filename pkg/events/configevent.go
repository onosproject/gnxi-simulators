// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package events

import "time"

// ConfigEvent a configuration event
type ConfigEvent EventHappend

// Clone clones the Event
func (ce *ConfigEvent) Clone() Event {
	clone := &ConfigEvent{}
	clone.Etype = ce.Etype
	clone.Subject = ce.Subject
	clone.Time = ce.Time
	clone.Values = ce.Values
	clone.Client = ce.Client
	return clone
}

// GetType returns type of an Event
func (ce *ConfigEvent) GetType() EventType {
	return ce.Etype
}

// GetTime returns the time when the event occurs
func (ce *ConfigEvent) GetTime() time.Time {
	return ce.Time
}

// GetValues returns the values of the event
func (ce *ConfigEvent) GetValues() interface{} {
	return ce.Values
}

// GetSubject returns the subject of the event
func (ce *ConfigEvent) GetSubject() string {
	return ce.Subject
}

// GetClient returns the stream client corresponding to the given ConfigEvent
func (ce *ConfigEvent) GetClient() interface{} {
	return ce.Client
}
