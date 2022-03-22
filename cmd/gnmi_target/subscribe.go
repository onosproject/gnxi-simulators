// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/google/gnxi/utils/credentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// Subscribe overrides the Subscribe function of gnmi.Target to provide user auth.
func (s *server) Subscribe(stream pb.GNMI_SubscribeServer) error {
	msg, ok := credentials.AuthorizeUser(stream.Context())
	if !ok {
		log.Infof("denied a Subscribe request: %v", msg)
		return status.Error(codes.PermissionDenied, msg)
	}

	return s.Server.Subscribe(stream)
}
