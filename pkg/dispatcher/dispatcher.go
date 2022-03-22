// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package dispatcher

import (
	"reflect"
	"sync"

	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/onosproject/gnxi-simulators/pkg/events"
)

var log = logging.GetLogger("dispatcher")

//Dispatcher dispatches the events
type Dispatcher struct {
	handlers map[reflect.Type][]reflect.Value
	lock     *sync.RWMutex
}

// NewDispatcher creates an instance of Dispatcher struct
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[reflect.Type][]reflect.Value),
		lock:     &sync.RWMutex{},
	}
}

// RegisterEvent registers custom events making it possible to register listeners for them.
func (d *Dispatcher) RegisterEvent(event events.Event) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	typ := reflect.TypeOf(event).Elem()
	log.Info("Registering the ", typ)
	if _, ok := d.handlers[typ]; ok {
		return false
	}
	var chanArr []reflect.Value
	d.handlers[typ] = chanArr
	return true
}

// RegisterListener registers chanel accepting desired event - a listener.
func (d *Dispatcher) RegisterListener(pipe interface{}) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	channelValue := reflect.ValueOf(pipe)
	channelType := channelValue.Type()
	if channelType.Kind() != reflect.Chan {
		panic("Trying to register a non-channel listener")
	}
	channelIn := channelType.Elem()
	if arr, ok := d.handlers[channelIn]; ok {
		d.handlers[channelIn] = append(arr, channelValue)
		return true
	}
	return false
}

// Dispatch provides thread safe method to send event to all listeners
// Returns true if succeeded and false if event was not registered
func (d *Dispatcher) Dispatch(event events.Event) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()

	eventType := reflect.TypeOf(event).Elem()
	if listeners, ok := d.handlers[eventType]; ok {
		for _, listener := range listeners {
			listener.TrySend(reflect.ValueOf(event.Clone()).Elem())
		}
		return true
	}
	return false
}
