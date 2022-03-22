// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package gnmi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/proto"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// SetDateTime update current-datetime field in runtime
func (s *Server) SetDateTime() error {
	s.configMu.Lock()
	defer s.configMu.Unlock()
	var path pb.Path
	textPbPath := `elem:<name:"system" > elem:<name:"state" > elem:<name:"current-datetime" > `
	if err := proto.UnmarshalText(textPbPath, &path); err != nil {
		return err
	}

	val := &pb.TypedValue{
		Value: &pb.TypedValue_StringVal{
			StringVal: time.Now().Format("2006-01-02T15:04:05Z-07:00"),
		},
	}
	update := &pb.Update{Path: &path, Val: val}

	jsonTree, _ := ygot.ConstructIETFJSON(s.config, &ygot.RFC7951JSONConfig{})
	_, _ = s.doReplaceOrUpdate(jsonTree, pb.UpdateResult_UPDATE, nil, update.GetPath(), update.GetVal())
	jsonDump, err := json.Marshal(jsonTree)
	if err != nil {
		msg := fmt.Sprintf("error in marshaling IETF JSON tree to bytes: %v", err)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}
	rootStruct, err := s.model.NewConfigStruct(jsonDump)
	if err != nil {
		msg := fmt.Sprintf("error in creating config struct from IETF JSON data: %v", err)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}
	s.config = rootStruct
	s.ConfigUpdate.In() <- update
	return nil

}
