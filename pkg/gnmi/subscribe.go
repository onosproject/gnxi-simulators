// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

// processSubStreamOnChange processes subscribe stream requests for on_change subscription mode.
func (s *Server) processSubStreamOnChange(c *streamClient, request *pb.SubscriptionList) {
	go s.listenToConfigEvents(request)

}

// processSubStreamSample processes subscribe stream requests for sample subscription mode.
func (s *Server) processSubStreamSample(c *streamClient, request *pb.SubscriptionList) {
	ticker := time.NewTicker(time.Duration(c.sampleInterval) * time.Nanosecond)
	go func() {
		for range ticker.C {
			s.collector(c, request)
		}
	}()
	s.listenForUpdates(c)

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
			// Adds streamClient to the list of subscribers
			for _, sub := range subscribe.Subscription {
				s.addSubscriber(sub.GetPath().String(), &c)
			}

			for _, sub := range subscribe.Subscription {
				switch sub.GetMode() {
				case pb.SubscriptionMode_ON_CHANGE:
					go s.processSubStreamOnChange(&c, subscribe)
				case pb.SubscriptionMode_SAMPLE:
					subSampleInterval := sub.GetSampleInterval()
					//If the sample_interval is set to 0,
					// the target MUST create the subscription and send the data with the
					// lowest interval possible for the target.
					if subSampleInterval == 0 {
						c.sampleInterval = lowestSampleInterval
					} else {
						// We assume that the target cannot support
						// the sample interval less than the lowest
						// sample interval which is defined in the target
						if subSampleInterval < lowestSampleInterval {
							return status.Error(codes.InvalidArgument, fmt.Sprintf("%s%d", "The sample interval must be higher than ", lowestSampleInterval))
						}
						c.sampleInterval = subSampleInterval

					}
					go s.processSubStreamSample(&c, subscribe)
				case pb.SubscriptionMode_TARGET_DEFINED:
					// TODO: when a client creates a
					// subscription specifying the target defined mode,
					// the target MUST determine the best type of subscription to
					// be created on a per-leaf basis.
					// That is to say,
					// if the path specified within the message refers
					//  to some leaves which are event driven
					// (e.g., the changing of state of an entity based on an external trigger)
					//  then an ON_CHANGE subscription may be created,
					// whereas if other data represents counter values,
					// a SAMPLE subscription may be created.
					go s.processSubStreamOnChange(&c, subscribe)

				}

			}

		default:
		}
	}

}
