// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"github.com/eapache/channels"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("gnmi")

// NewServer creates an instance of Server with given json config.
func NewServer(model *Model, config []byte, callback ConfigCallback) (*Server, error) {
	rootStruct, err := model.NewConfigStruct(config)
	if err != nil {
		return nil, err
	}
	s := &Server{
		model:    model,
		config:   rootStruct,
		callback: callback,
	}
	if config != nil && s.callback != nil {
		if err := s.callback(rootStruct); err != nil {
			return nil, err
		}
	}
	// Initialize readOnlyUpdateValue variable

	val := &pb.TypedValue{
		Value: &pb.TypedValue_StringVal{
			StringVal: "INIT_STATE",
		},
	}
	s.readOnlyUpdateValue = &pb.Update{Path: nil, Val: val}
	s.subscribers = make(map[string]*streamClient)
	s.ConfigUpdate = channels.NewRingChannel(100)

	return s, nil
}
