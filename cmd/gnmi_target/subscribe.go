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

package main

import (
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/onosproject/simulators/pkg/dispatcher"
	"github.com/onosproject/simulators/pkg/events"

	"github.com/openconfig/gnmi/proto/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// randomEventProducer produces update events for stream subscribers
func (s *server) randomEventProducer(dispatcher *dispatcher.Dispatcher,
	subscribe *pb.SubscriptionList) {
	for {
		for _, sub := range subscribe.Subscription {

			ipPrefix := "192.168.1"
			ipSuffix := strconv.Itoa(rand.Intn(254))
			ip := ipPrefix + "." + ipSuffix
			subject := "subscribe_stream_randm_event"
			val := &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: ip,
				},
			}
			update, _ := s.getUpdate(subscribe, sub.GetPath())
			update.Val = val

			event := &events.RandomEvent{
				Subject: subject,
				Time:    time.Now(),
				Etype:   events.EventTypeRandom,
				Values:  update,
			}
			dispatcher.Dispatch(event)
			time.Sleep(randomEventInterval)
		}

	}
}

// configEventProducer produces update events for stream subscribers
func (s *server) configEventProducer(dispatcher *dispatcher.Dispatcher,
	subscribe *pb.SubscriptionList) {
	for update := range s.UpdateChann {
		for _, sub := range subscribe.Subscription {

			if reflect.DeepEqual(sub.GetPath().String(), update.GetPath().String()) {
				newValue, _ := s.getUpdate(subscribe, update.GetPath())
				subject := "subscribe_stream"
				update.Val = newValue.Val

				event := &events.ConfigEvent{
					Subject: subject,
					Time:    time.Now(),
					Etype:   events.EventTypeConfiguration,
					Values:  update,
				}
				dispatcher.Dispatch(event)
			}

		}

	}
}

// sendRandomEvent stream random events to the subscribed clients.
// This function is just for testing purposes and is not part of the
// gnmi specification.
func (s *server) sendRandomEvent(subscribe *gnmi.SubscriptionList,
	stream pb.GNMI_SubscribeServer) {
	dispatcher := dispatcher.NewDispatcher()
	ok := dispatcher.RegisterEvent((*events.RandomEvent)(nil))

	if !ok {
		log.Error("Cannot register an event")
	}

	ch := make(chan events.RandomEvent, 100)
	ok = dispatcher.RegisterListener(ch)

	if !ok {
		log.Error("Cannot register the listener")
	}
	go s.randomEventProducer(dispatcher, subscribe)
	for result := range ch {

		var update *pb.Update
		update = result.GetValues().(*pb.Update)

		response, _ := buildSubResponse(update)
		// Update the readOnlyUpdateValue variable to be accessible with get function
		s.readOnlyUpdateValue = update

		s.sendResponse(response, stream)
		responseSync := &pb.SubscribeResponse_SyncResponse{
			SyncResponse: true,
		}
		response = &pb.SubscribeResponse{
			Response: responseSync,
		}

		s.sendResponse(response, stream)

	}
}

// sendConfigEvent sends a config event to the subscribers
func (s *server) sendConfigEvent(subscribe *gnmi.SubscriptionList,
	stream pb.GNMI_SubscribeServer) {
	dispatcher := dispatcher.NewDispatcher()
	ok := dispatcher.RegisterEvent((*events.ConfigEvent)(nil))

	if !ok {
		log.Error("Cannot register an event")
	}

	ch := make(chan events.ConfigEvent, 100)
	ok = dispatcher.RegisterListener(ch)

	if !ok {
		log.Error("Cannot register the listener")
	}
	go s.configEventProducer(dispatcher, subscribe)
	for result := range ch {

		var update *pb.Update
		update = result.GetValues().(*pb.Update)

		response, _ := buildSubResponse(update)

		s.sendResponse(response, stream)
		responseSync := &pb.SubscribeResponse_SyncResponse{
			SyncResponse: true,
		}
		response = &pb.SubscribeResponse{
			Response: responseSync,
		}
		s.sendResponse(response, stream)

	}

}

// sendStreamResults stream updates to the subscribed clients.
func (s *server) sendStreamResults(subscribe *gnmi.SubscriptionList,
	stream pb.GNMI_SubscribeServer) {

	for _, sub := range subscribe.Subscription {
		if strings.Compare(sub.GetPath().String(), readOnlyPath) == 0 {
			s.sendRandomEvent(subscribe, stream)
		}
	}
	s.sendConfigEvent(subscribe, stream)

}

// Subscribe overrides the Subscribe function to implement it.
func (s *server) Subscribe(stream pb.GNMI_SubscribeServer) error {
	c := streamClient{stream: stream}
	var err error
	updateChan := make(chan *pb.Update)
	var subscribe *pb.SubscriptionList
	var mode gnmi.SubscriptionList_Mode

	for {
		c.sr, err = stream.Recv()
		switch {
		case err == io.EOF:
			log.Error("No more input is available, subscription terminated")
			return nil
		case err != nil:
			log.Error("Error in subscription", err)
			return err

		}

		if c.sr.GetPoll() != nil {
			mode = gnmi.SubscriptionList_POLL
		} else {
			subscribe = c.sr.GetSubscribe()
			mode = subscribe.Mode
		}
		done := make(chan struct{})

		//If the subscription mode is ONCE or POLL we immediately start a routine to collect the data
		if mode != pb.SubscriptionList_STREAM {
			go s.collector(updateChan, subscribe)
		}
		go s.listenForUpdates(updateChan, stream, mode, done)

		if mode == pb.SubscriptionList_ONCE {
			<-done
			return nil
		} else if mode == pb.SubscriptionList_STREAM {
			s.sendStreamResults(subscribe, stream)
			return nil
		}

	}

}
