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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var (
	idPattern = `[a-zA-Z_][a-zA-Z\d\_\-\.]*`
	// YANG identifiers must follow RFC 6020:
	// https://tools.ietf.org/html/rfc6020#section-6.2.
	idRe = regexp.MustCompile(`^` + idPattern + `$`)
	// The sting representation of List key value pairs must follow the
	// following pattern: [key=value], where key is the List key leaf name,
	// and value is the string representation of key leaf value.
	kvRe = regexp.MustCompile(`^\[` +
		// Key leaf name must be a valid YANG identifier.
		idPattern + `=` +
		// Key leaf value must be a non-empty string, which may contain
		// newlines. Use (?s) to turn on s flag to match newlines.
		`((?s).+)` +
		`\]$`)
)

// splitPath splits a string representation of path into []string. Path
// elements are separated by '/'. String splitting scans from left to right. A
// '[' marks the beginning of a List key value pair substring. A List key value
// pair string ends at the first ']' encountered. Neither an escaped '[', i.e.,
// `\[`, nor an escaped ']', i.e., `\]`, serves as the boundary of a List key
// value pair string.
//
// Within a List key value string, '/', '[' and ']' are treated differently:
//
//	1. A '/' does not act as a separator, and is allowed to be part of a
//	List key leaf value.
//
//	2. A '[' is allowed within a List key value. '[' and `\[` are
//	equivalent within a List key value.
//
//	3. If a ']' needs to be part of a List key value, it must be escaped as
//	'\]'. The first unescaped ']' terminates a List key value string.
//
// Outside of any List key value pair string:
//
//	1. A ']' without a matching '[' does not generate any error in this
//	API. This error is caught later by another API.
//
//	2. A '[' without an closing ']' is treated as an error, because it
//	indicates an incomplete List key leaf value string.
//
// For example, "/a/b/c" is split into []string{"a", "b", "c"}.
// "/a/b[k=eth1/1]/c" is split into []string{"a", "b[k=eth1/1]", "c"}.
// `/a/b/[k=v\]]/c` is split into []string{"a", "b", `[k=v\]]`, "c"}.
// "a/b][k=v]/c" is split into []string{"a", "b][k=v]", "c"}. The invalid List
// name "b]" error will be caught later by another API. "/a/b[k=v/c" generates
// an error because of incomplete List key value pair string.
func splitPath(str string) ([]string, error) {
	var path []string
	str += "/"
	// insideBrackets is true when at least one '[' has been found and no
	// ']' has been found. It is false when a closing ']' has been found.
	insideBrackets := false
	// begin marks the beginning of a path element, which is separated by
	// '/' unclosed between '[' and ']'.
	begin := 0
	// end marks the end of a path element, which is separated by '/'
	// unclosed between '[' and ']'.
	end := 0

	// Split the given string using unescaped '/'.
	for end < len(str) {
		switch str[end] {
		case '/':
			if !insideBrackets {
				// Current '/' is a valid path element
				// separator.
				if end > begin {
					path = append(path, str[begin:end])
				}
				end++
				begin = end
			} else {
				// Current '/' must be part of a List key value
				// string.
				end++
			}
		case '[':
			if (end == 0 || str[end-1] != '\\') && !insideBrackets {
				// Current '[' is unescacped, and is the
				// beginning of List key-value pair(s) string.
				insideBrackets = true
			}
			end++
		case ']':
			if (end == 0 || str[end-1] != '\\') && insideBrackets {
				// Current ']' is unescacped, and is the end of
				// List key-value pair(s) string.
				insideBrackets = false
			}
			end++
		default:
			end++
		}
	}

	if insideBrackets {
		return nil, fmt.Errorf("missing ] in path string: %s", str)
	}
	return path, nil
}

// parseKeyValueString parses a List key-value pair, and returns a
// map[string]string whose key is the List key leaf name and whose value is the
// string representation of List key leaf value. The input path-valur pairs are
// encoded using the following pattern: [k1=v1][k2=v2]..., where k1 and k2 must be
// valid YANG identifiers, v1 and v2 can be any non-empty strings where any ']'
// must be escapced by an '\'. Any malformed key-value pair generates an error.
// For example, given
//	"[k1=v1][k2=v2]",
// this API returns
//	map[string]string{"k1": "v1", "k2": "v2"}.
func parseKeyValueString(str string) (map[string]string, error) {
	keyValuePairs := make(map[string]string)
	// begin marks the beginning of a key-value pair.
	begin := 0
	// end marks the end of a key-value pair.
	end := 0
	// insideBrackets is true when at least one '[' has been found and no
	// ']' has been found. It is false when a closing ']' has been found.
	insideBrackets := false

	for end < len(str) {
		switch str[end] {
		case '[':
			if (end == 0 || str[end-1] != '\\') && !insideBrackets {
				insideBrackets = true
			}
			end++
		case ']':
			if (end == 0 || str[end-1] != '\\') && insideBrackets {
				insideBrackets = false
				keyValue := str[begin : end+1]
				// Key-value pair string must have the
				// following pattern: [k=v], where k is a valid
				// YANG identifier, and v can be any non-empty
				// string.
				if !kvRe.MatchString(keyValue) {
					return nil, fmt.Errorf("malformed List key-value pair string: %s, in: %s", keyValue, str)
				}
				keyValue = keyValue[1 : len(keyValue)-1]
				i := strings.Index(keyValue, "=")
				key, val := keyValue[:i], keyValue[i+1:]
				// Recover escaped '[' and ']'.
				val = strings.Replace(val, `\]`, `]`, -1)
				val = strings.Replace(val, `\[`, `[`, -1)
				keyValuePairs[key] = val
				begin = end + 1
			}
			end++
		default:
			end++
		}
	}

	if begin < end {
		return nil, fmt.Errorf("malformed List key-value pair string: %s", str)
	}

	return keyValuePairs, nil
}

// parseElement parses a split path element, and returns the parsed elements.
// Two types of path elements are supported:
//
// 1. Non-List schema node names which must be valid YANG identifiers. A valid
// schema node name is returned as it is. For example, given "abc", this API
// returns []interface{"abc"}.
//
// 2. List elements following this pattern: list-name[k1=v1], where list-name
// is the substring from the beginning of the input string to the first '[', k1
// is the substring from the letter after '[' to the first '=', and v1 is the
// substring from the letter after '=' to the first unescaped ']'. list-name
// and k1 must be valid YANG identifier, and v1 can be any non-empty string
// where ']' is escaped by '\'. A List element is parsed into two parts: List
// name and List key value pair(s). List key value pairs are saved in a
// map[string]string whose key is List key leaf name and whose value is the
// string representation of List key leaf value. For example, given
//	"list-name[k1=v1]",
// this API returns
//	[]interface{}{"list-name", map[string]string{"k1": "v1"}}.
// Multi-key List elements follow a similar pattern:
//	list-name[k1=v1]...[kN=vN].
func parseElement(elem string) ([]interface{}, error) {
	i := strings.Index(elem, "[")
	if i < 0 {
		if !idRe.MatchString(elem) {
			return nil, fmt.Errorf("invalid node name: %q", elem)
		}
		return []interface{}{elem}, nil
	}

	listName := elem[:i]
	if !idRe.MatchString(listName) {
		return nil, fmt.Errorf("invalid List name: %q, in: %s", listName, elem)
	}
	keyValuePairs, err := parseKeyValueString(elem[i:])
	if err != nil {
		return nil, fmt.Errorf("invalid path element %s: %v", elem, err)
	}
	return []interface{}{listName, keyValuePairs}, nil
}

// ParseStringPath parses a string path and produces a []interface{} of parsed
// path elements. Path elements in a string path are separated by '/'. Each
// path element can either be a schema node name or a List path element. Schema
// node names must be valid YANG identifiers. A List path element is encoded
// using the following pattern: list-name[key1=value1]...[keyN=valueN]. Each
// List path element generates two parsed path elements: List name and a
// map[string]string containing List key-value pairs with value(s) in string
// representation. A '/' within a List key value pair string, i.e., between a
// pair of '[' and ']', does not serve as a path separator, and is allowed to be
// part of a List key leaf value. For example, given a string path:
//	"/a/list-name[k=v/v]/c",
// this API returns:
//	[]interface{}{"a", "list-name", map[string]string{"k": "v/v"}, "c"}.
//
// String path parsing consists of two passes. In the first pass, the input
// string is split into []string using valid separator '/'. An incomplete List
// key value string, i.e, a '[' which starts a List key value string without a
// closing ']', in input string generates an error. In the above example, this
// pass produces:
//	[]string{"a", "list-name[k=v/v]", "c"}.
// In the second pass, each element in split []string is parsed checking syntax
// and pattern correctness. Errors are generated for invalid YANG identifiers,
// malformed List key-value string, etc.. In the above example, the second pass
// produces:
//	[]interface{}{"a", "list-name", map[string]string{"k", "v/v"}, "c"}.
func ParseStringPath(stringPath string) ([]interface{}, error) {
	elems, err := splitPath(stringPath)
	if err != nil {
		return nil, err
	}

	var path []interface{}
	// Check whether each path element is valid. Parse List key value
	// pairs.
	for _, elem := range elems {
		parts, err := parseElement(elem)
		if err != nil {
			return nil, fmt.Errorf("invalid string path %s: %v", stringPath, err)
		}
		path = append(path, parts...)
	}

	return path, nil
}

// getChildNode gets a node's child with corresponding schema specified by path
// element. If not found and createIfNotExist is set as true, an empty node is
// created and returned.
func getChildNode(node map[string]interface{}, schema *yang.Entry, elem *pb.PathElem, createIfNotExist bool) (interface{}, *yang.Entry) {
	var nextSchema *yang.Entry
	var ok bool

	if nextSchema, ok = schema.Dir[elem.Name]; !ok {
		return nil, nil
	}

	var nextNode interface{}
	if elem.GetKey() == nil {
		if nextNode, ok = node[elem.Name]; !ok {
			if createIfNotExist {
				node[elem.Name] = make(map[string]interface{})
				nextNode = node[elem.Name]
			}
		}
		return nextNode, nextSchema
	}

	nextNode = getKeyedListEntry(node, elem, createIfNotExist)
	return nextNode, nextSchema
}

// getKeyedListEntry finds the keyed list entry in node by the name and key of
// path elem. If entry is not found and createIfNotExist is true, an empty entry
// will be created (the list will be created if necessary).
func getKeyedListEntry(node map[string]interface{}, elem *pb.PathElem, createIfNotExist bool) map[string]interface{} {
	curNode, ok := node[elem.Name]
	if !ok {
		if !createIfNotExist {
			return nil
		}

		// Create a keyed list as node child and initialize an entry.
		m := make(map[string]interface{})
		for k, v := range elem.Key {
			m[k] = v
			if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
				m[k] = vAsNum
			}
		}
		node[elem.Name] = []interface{}{m}
		return m
	}

	// Search entry in keyed list.
	keyedList, ok := curNode.([]interface{})
	if !ok {
		return nil
	}
	for _, n := range keyedList {
		m, ok := n.(map[string]interface{})
		if !ok {
			log.Errorf("wrong keyed list entry type: %T", n)
			return nil
		}
		keyMatching := true
		// must be exactly match
		for k, v := range elem.Key {
			attrVal, ok := m[k]
			if !ok {
				return nil
			}
			if v != fmt.Sprintf("%v", attrVal) {
				keyMatching = false
				break
			}
		}
		if keyMatching {
			return m
		}
	}
	if !createIfNotExist {
		return nil
	}

	// Create an entry in keyed list.
	m := make(map[string]interface{})
	for k, v := range elem.Key {
		m[k] = v
		if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
			m[k] = vAsNum
		}
	}
	node[elem.Name] = append(keyedList, m)
	return m
}

// doDelete deletes the path from the json tree if the path exists.
func (s *Server) doDeletePath(jsonTree map[string]interface{}, prefix, path *pb.Path) (bool, error) {
	// Update json tree of the device config
	fullPath := gnmiFullPath(prefix, path)
	var curNode interface{} = jsonTree
	pathDeleted := false
	schema := s.model.schemaTreeRoot
	for i, elem := range fullPath.Elem { // Delete sub-tree or leaf node.
		log.Info("index: ", i, elem)
		node, ok := curNode.(map[string]interface{})
		//log.Info("node ", node)
		if !ok {
			log.Info("break")
			break
		}

		// Delete node
		log.Info("i, and full path length:", i, len(fullPath.Elem)-1)
		if i == len(fullPath.Elem)-1 {
			log.Info("here")
			if elem.GetKey() == nil {
				log.Info("here 2", elem.Name)
				log.Info("node: ", node)
				delete(node, elem.GetName())
				log.Info("node after delete:", node)
				pathDeleted = true
				break
			}
			pathDeleted = deleteKeyedListEntry(node, elem)
			break
		}

		if curNode, schema = getChildNode(node, schema, elem, false); curNode == nil {
			log.Info("break 2")
			break
		}
	}

	if pathDeleted == false {
		return pathDeleted, nil
	}

	if reflect.DeepEqual(fullPath, pbRootPath) { // Delete root
		for k := range jsonTree {
			log.Info("k:", k)
			delete(jsonTree, k)
		}
	}

	return pathDeleted, nil

}

func (s *Server) parseArray(array []interface{}, path []string, pathFlag bool, dataType string) {
	for _, val := range array {
		switch val.(type) {
		case map[string]interface{}:
			s.ParseJSONTree(val.(map[string]interface{}), path, pathFlag, dataType)
		case []interface{}:
			s.parseArray(val.([]interface{}), path, pathFlag, dataType)

		}
	}
}

// ParseJSONTree parses a nested map based on the given type for the get request
func (s *Server) ParseJSONTree(aMap map[string]interface{}, path []string, pathFlag bool, dataType string) error {

	for key, val := range aMap {
		switch val.(type) {
		case map[string]interface{}:

			//log.Info("key1:", key)
			path = append(path, key)
			if strings.Compare(key, dataType) == 0 {
				pathFlag = true
			}
			s.ParseJSONTree(val.(map[string]interface{}), path, pathFlag, dataType)
			//log.Info("path1:", path)
			//path = ""

		case []interface{}:

			//log.Info("key2:", key, ":", val)
			path = append(path, key)
			if strings.Compare(key, dataType) == 0 {
				pathFlag = true
			}
			s.parseArray(val.([]interface{}), path, pathFlag, dataType)

			//log.Info("   ", path)
			//path = ""
		default:
			//log.Info("   ", key, ":", val)
			//log.Info(path)
			//if pathFlag == false {
			path = path[:len(path)-1]
			log.Info(path)

			//}
			//path = ""
			break
		}

	}

	return nil
}

// ToGNMIPath parses an xpath string into a gnmi Path struct defined in gnmi
// proto. Path convention can be found in
// https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-path-conventions.md
//
// For example, xpath /interfaces/interface[name=Ethernet1/2/3]/state/counters
// will be parsed to:
//
//    elem: <name: "interfaces" >
//    elem: <
//        name: "interface"
//        key: <
//            key: "name"
//            value: "Ethernet1/2/3"
//        >
//    >
//    elem: <name: "state" >
//    elem: <name: "counters" >
func ToGNMIPath(xpath string) (*pb.Path, error) {
	xpathElements, err := ParseStringPath(xpath)
	if err != nil {
		return nil, err
	}
	var pbPathElements []*pb.PathElem
	for _, elem := range xpathElements {
		switch v := elem.(type) {
		case string:
			pbPathElements = append(pbPathElements, &pb.PathElem{Name: v})
		case map[string]string:
			n := len(pbPathElements)
			if n == 0 {
				return nil, fmt.Errorf("missing name before key-value list")
			}
			if pbPathElements[n-1].Key != nil {
				return nil, fmt.Errorf("two subsequent key-value lists")
			}
			pbPathElements[n-1].Key = v
		default:
			return nil, fmt.Errorf("wrong data type: %T", v)
		}
	}
	return &pb.Path{Elem: pbPathElements}, nil
}

// isNIl checks if an interface is nil or its value is nil.
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch kind := reflect.ValueOf(i).Kind(); kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}

func (s *Server) toGoStruct(jsonTree map[string]interface{}) (ygot.ValidatedGoStruct, error) {
	jsonDump, err := json.Marshal(jsonTree)
	if err != nil {
		return nil, fmt.Errorf("error in marshaling IETF JSON tree to bytes: %v", err)
	}
	goStruct, err := s.model.NewConfigStruct(jsonDump)
	if err != nil {
		return nil, fmt.Errorf("error in creating config struct from IETF JSON data: %v", err)
	}
	return goStruct, nil
}

// getGNMIServiceVersion returns a pointer to the gNMI service version string.
// The method is non-trivial because of the way it is defined in the proto file.
func getGNMIServiceVersion() (*string, error) {
	gzB, _ := (&pb.Update{}).Descriptor()
	r, err := gzip.NewReader(bytes.NewReader(gzB))
	if err != nil {
		return nil, fmt.Errorf("error in initializing gzip reader: %v", err)
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error in reading gzip data: %v", err)
	}
	desc := &dpb.FileDescriptorProto{}
	if err := proto.Unmarshal(b, desc); err != nil {
		return nil, fmt.Errorf("error in unmarshaling proto: %v", err)
	}
	ver, err := proto.GetExtension(desc.Options, pb.E_GnmiService)
	if err != nil {
		return nil, fmt.Errorf("error in getting version from proto extension: %v", err)
	}
	return ver.(*string), nil
}

// gnmiFullPath builds the full path from the prefix and path.
func gnmiFullPath(prefix, path *pb.Path) *pb.Path {
	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElement() != nil {
		fullPath.Element = append(prefix.GetElement(), path.GetElement()...)
	}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}

// checkEncodingAndModel checks whether encoding and models are supported by the server. Return error if anything is unsupported.
func (s *Server) checkEncodingAndModel(encoding pb.Encoding, models []*pb.ModelData) error {
	hasSupportedEncoding := false
	for _, supportedEncoding := range supportedEncodings {
		if encoding == supportedEncoding {
			hasSupportedEncoding = true
			break
		}
	}
	if !hasSupportedEncoding {
		return fmt.Errorf("unsupported encoding: %s", pb.Encoding_name[int32(encoding)])
	}
	for _, m := range models {
		isSupported := false
		for _, supportedModel := range s.model.modelData {
			if reflect.DeepEqual(m, supportedModel) {
				isSupported = true
				break
			}
		}
		if !isSupported {
			return fmt.Errorf("unsupported model: %v", m)
		}
	}
	return nil
}
