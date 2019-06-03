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

	for i, path := range paths {
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}
		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}

		node, stat := ygotutils.GetNode(s.model.schemaTreeRoot, s.config, fullPath)
		if isNil(node) || stat.GetCode() != int32(cpb.Code_OK) {
			return nil, status.Errorf(codes.NotFound, "path %v not found", fullPath)
		}

		ts := time.Now().UnixNano()

		nodeStruct, ok := node.(ygot.GoStruct)
		// Return leaf node.
		dataTypeFlag := false
		if !ok {
			// check the requested leaf value and its dataType match with each other
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

		elements := fullPath.GetElem()
		dataTypeString := strings.ToLower(dataType.String())
		for _, elem := range elements {

			if strings.Compare(dataTypeString, elem.GetName()) == 0 {
				dataTypeFlag = true
			}
			if strings.Compare(dataTypeString, "all") == 0 {
				dataTypeFlag = true
			}
		}

		if req.GetUseModels() != nil {
			return nil, status.Errorf(codes.Unimplemented, "filtering Get using use_models is unsupported, got: %v", req.GetUseModels())
		}

		// Return IETF JSON by default.
		jsonEncoder := func() (map[string]interface{}, error) {
			return ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: true})
		}
		jsonType := "IETF"
		buildUpdate := func(b []byte) *pb.Update {
			return &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonIetfVal{JsonIetfVal: b}}}
		}

		if req.GetEncoding() == pb.Encoding_JSON {
			jsonEncoder = func() (map[string]interface{}, error) {
				return ygot.ConstructInternalJSON(nodeStruct)
			}
			jsonType = "Internal"
			buildUpdate = func(b []byte) *pb.Update {
				return &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonVal{JsonVal: b}}}
			}
		}

		var jsonTree map[string]interface{}
		var err error
		jsonTree, err = jsonEncoder()

		log.Info(jsonTree)

		//if dataTypeFlag == false {

		fullPathElements := fullPath.GetElem()
		updatedMap := make(map[string]interface{})
		root := fullPathElements[len(fullPathElements)-1]
		updatedMap[root.GetName()] = jsonTree
		for i := len(fullPathElements) - 2; i >= 0; i-- {
			log.Info("index, elem: ", i, fullPathElements[i])

			tempMap := make(map[string]interface{})
			tempMap[fullPathElements[i].GetName()] = updatedMap
			updatedMap = tempMap

		}
		var pathString []string
		s.ParseJSONTree(updatedMap, pathString, false, dataTypeString)
		/*log.Info("updatedMAp:", updatedMap)
		newPath1 := "/system/openflow/"
		removedPath1, _ := ToGNMIPath(newPath1)
		//log.Info("The path that should be removed ", removedPath1)

		s.doDeletePath(updatedMap, prefix, removedPath1)

		log.Info(updatedMap)

		newPath2 := "/system/clock/"
		removedPath2, _ := ToGNMIPath(newPath2)
		//log.Info("The path that should be removed ", removedPath2)

		s.doDeletePath(updatedMap, prefix, removedPath2)

		log.Info("result:", updatedMap)

		newPath3 := "/system/config/"
		removedPath3, _ := ToGNMIPath(newPath3)
		//log.Info("The path that should be removed ", removedPath2)

		s.doDeletePath(updatedMap, prefix, removedPath3)

		log.Info("result:", updatedMap)
		//}*/

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

		update := buildUpdate(jsonDump)
		notifications[i] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
	}
	resp := &pb.GetResponse{Notification: notifications}

	return resp, nil
}
