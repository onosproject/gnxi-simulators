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

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"io"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// processSubscribeOnce processes subscribe once requests
func (s *Server) processSubscribeOnce(c *streamClient, request *pb.SubscriptionList) {
	go s.collector(c, request)
	s.listenForUpdates(c)

}

// processSubscribePoll processes subcribe poll requests
func (s *Server) processSubscribePoll(c *streamClient, request *pb.SubscriptionList) {
	go s.collector(c, request)
	s.listenForUpdates(c)
}

// processSubscribeStream processes subscribe stream requests.
func (s *Server) processSubscribeStream(c *streamClient, request *pb.SubscriptionList) {
	go s.listenToConfigEvents(request)
}

// Subscribe handle subscribe requests including POLL, STREAM, ONCE subscribe requests
func (s *Server) Subscribe(stream pb.GNMI_SubscribeServer) error {

	c := streamClient{stream: stream}
	var err error
	c.UpdateChan = make(chan *pb.Update, 100)

	var subscribe *pb.SubscriptionList
	var mode gnmi.SubscriptionList_Mode

	for {
		c.sr, err = stream.Recv()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		}

		if c.sr.GetPoll() != nil {
			mode = gnmi.SubscriptionList_POLL
		} else {
			subscribe = c.sr.GetSubscribe()
			mode = subscribe.Mode
		}

		switch mode {
		case pb.SubscriptionList_ONCE:
			go s.processSubscribeOnce(&c, subscribe)
		case pb.SubscriptionList_POLL:
			go s.processSubscribePoll(&c, subscribe)
		case pb.SubscriptionList_STREAM:
			for _, sub := range subscribe.Subscription {
				if strings.Compare(sub.GetPath().String(), readOnlyPath) == 0 {
					go s.sendRandomEvent(&c, subscribe)
				}
			}
			// Adds streamClient to the list of subscribers
			for _, sub := range subscribe.Subscription {
				s.subscribers[sub.GetPath().String()] = &c
			}
			go s.processSubscribeStream(&c, subscribe)

		default:
		}
	}

}
