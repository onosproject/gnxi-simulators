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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/golang/glog"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/experimental/ygotutils"
	"github.com/openconfig/ygot/ygot"
	"golang.org/x/net/context"
	cpb "google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get implements the Get RPC in gNMI spec.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	dataType := req.GetType()

	if err := s.checkEncodingAndModel(req.GetEncoding(), req.GetUseModels()); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*pb.Notification, len(paths))

	s.mu.RLock()
	defer s.mu.RUnlock()

	if paths == nil && dataType.String() != "" {

		jsonType := "IETF"
		if req.GetEncoding() == pb.Encoding_JSON {
			jsonType = "Internal"
		}
		notifications := make([]*pb.Notification, 1)
		path := pb.Path{}
		// Gets the whole config data tree
		node, stat := ygotutils.GetNode(s.model.schemaTreeRoot, s.config, &path)
		if isNil(node) || stat.GetCode() != int32(cpb.Code_OK) {
			return nil, status.Errorf(codes.NotFound, "path %v not found", path)
		}

		nodeStruct, _ := node.(ygot.GoStruct)
		jsonTree, _ := ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: true})

		jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataType.String()), &path).(map[string]interface{})
		jsonDump, err := json.Marshal(jsonTree)

		if err != nil {
			msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		ts := time.Now().UnixNano()

		update := buildUpdate(jsonDump, &path, jsonType)
		notifications[0] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
		resp := &pb.GetResponse{Notification: notifications}
		return resp, nil
	}

	for i, path := range paths {
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}

		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}

		ts := time.Now().UnixNano()
		// Handling the read random state field
		if strings.Compare(path.String(), readOnlyPath) == 0 {
			// If no subscribe request is initiated on random state variable
			// then we should return the initial value in the config otherwise
			// we will return the last random value which is stored in "readOnlyUpdateValue"
			// variable.
			if s.readOnlyUpdateValue.Val.GetStringVal() != "INIT_STATE" {
				update := s.readOnlyUpdateValue
				notifications[i] = &pb.Notification{
					Timestamp: ts,
					Prefix:    req.GetPrefix(),
					Update:    []*pb.Update{update},
				}
				resp := &pb.GetResponse{Notification: notifications}
				return resp, nil
			}
		}

		node, stat := ygotutils.GetNode(s.model.schemaTreeRoot, s.config, fullPath)
		if isNil(node) || stat.GetCode() != int32(cpb.Code_OK) {
			return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
		}

		ts = time.Now().UnixNano()

		nodeStruct, ok := node.(ygot.GoStruct)
		dataTypeFlag := false
		// Return leaf node.
		if !ok {
			elements := fullPath.GetElem()
			dataTypeString := strings.ToLower(dataType.String())
			if strings.Compare(dataTypeString, "all") == 0 {
				dataTypeFlag = true
			} else {
				for _, elem := range elements {
					if strings.Compare(dataTypeString, elem.GetName()) == 0 {
						dataTypeFlag = true
						break
					}

				}
			}
			if dataTypeFlag == false {
				return nil, status.Error(codes.Internal, "The requested dataType is not valid")
			}
			var val *pb.TypedValue
			switch kind := reflect.ValueOf(node).Kind(); kind {
			case reflect.Ptr, reflect.Interface:
				var err error
				val, err = value.FromScalar(reflect.ValueOf(node).Elem().Interface())
				if err != nil {
					msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
					log.Error(msg)
					return nil, status.Error(codes.Internal, msg)
				}
			case reflect.Int64:
				enumMap, ok := s.model.enumData[reflect.TypeOf(node).Name()]
				if !ok {
					return nil, status.Error(codes.Internal, "not a GoStruct enumeration type")
				}
				val = &pb.TypedValue{
					Value: &pb.TypedValue_StringVal{
						StringVal: enumMap[reflect.ValueOf(node).Int()].Name,
					},
				}
			default:
				return nil, status.Errorf(codes.Internal, "unexpected kind of leaf node type: %v %v", node, kind)
			}

			update := &pb.Update{Path: path, Val: val}
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}
		dataTypeString := strings.ToLower(dataType.String())

		if req.GetUseModels() != nil {
			return nil, status.Errorf(codes.Unimplemented, "filtering Get using use_models is unsupported, got: %v", req.GetUseModels())
		}

		jsonType := "IETF"

		if req.GetEncoding() == pb.Encoding_JSON {
			jsonType = "Internal"
		}

		var jsonTree map[string]interface{}
		var err error
		jsonTree, err = jsonEncoder(jsonType, nodeStruct)
		jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataTypeString), fullPath).(map[string]interface{})
		if err != nil {
			msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", jsonType, err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		jsonDump, err := json.Marshal(jsonTree)
		if err != nil {
			msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		update := buildUpdate(jsonDump, path, jsonType)
		notifications[i] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
	}
	resp := &pb.GetResponse{Notification: notifications}

	return resp, nil
}
