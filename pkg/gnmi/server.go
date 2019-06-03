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
	"sync"

	"github.com/openconfig/ygot/ygot"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// ConfigCallback is the signature of the function to apply a validated config to the physical device.
type ConfigCallback func(ygot.ValidatedGoStruct) error

var (
	pbRootPath         = &pb.Path{}
	supportedEncodings = []pb.Encoding{pb.Encoding_JSON, pb.Encoding_JSON_IETF}
)

// Server struct maintains the data structure for device config and implements the interface of gnmi server. It supports Capabilities, Get, and Set APIs.
// Typical usage:
//	g := grpc.NewServer()
//	s, err := Server.NewServer(model, config, callback)
//	pb.NewServer(g, s)
//	reflection.Register(g)
//	listen, err := net.Listen("tcp", ":8080")
//	g.Serve(listen)
//
// For a real device, apply the config changes to the hardware in the callback function.
// Arguments:
//		newConfig: new root config to be applied on the device.
// func callback(newConfig ygot.ValidatedGoStruct) error {
//		// Apply the config to your device and return nil if success. return error if fails.
//		//
//		// Do something ...
// }
type Server struct {
	model    *Model
	callback ConfigCallback

	config ygot.ValidatedGoStruct
	mu     sync.RWMutex // mu is the RW lock to protect the access to config
}

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
	return s, nil
}

// GetConfig returns the config store
func (s *Server) GetConfig() (ygot.ValidatedGoStruct, error) {
	return s.config, nil
}

// InternalUpdate is an experimental feature to let the server update its
// internal states. Use it with your own risk.
func (s *Server) InternalUpdate(fp func(config ygot.ValidatedGoStruct) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fp(s.config)
}
