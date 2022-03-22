// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Capabilities returns supported encodings and supported models.
func (s *Server) Capabilities(ctx context.Context, req *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
	ver, err := getGNMIServiceVersion()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in getting gnmi service version: %v", err)
	}
	return &pb.CapabilityResponse{
		SupportedModels:    s.model.modelData,
		SupportedEncodings: supportedEncodings,
		GNMIVersion:        ver,
	}, nil
}

// getGNMIServiceVersion returns a pointer to the gNMI service version string.
// The method is non-trivial because of the way it is defined in the proto file.
func getGNMIServiceVersion() (string, error) {
	parentFile := (&pb.Update{}).ProtoReflect().Descriptor().ParentFile()
	options := parentFile.Options()
	version := ""
	if fileOptions, ok := options.(*protobuf.FileOptions); ok {
		ver, err := proto.GetExtension(fileOptions, pb.E_GnmiService)
		if err != nil {
			return "", err
		}
		version = *ver.(*string)
	}
	return version, nil

}
