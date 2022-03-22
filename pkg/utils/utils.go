// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var leafValue string

// FindLeaf finds a leaf in the given JSON tree that is stored in a map.
func FindLeaf(aMap map[string]interface{}, leaf string) (string, error) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			FindLeaf(val.(map[string]interface{}), leaf)
		case []interface{}:
			ParseArray(val.([]interface{}), leaf)
		default:
			if leaf == key {
				leafValue = concreteVal.(string)
				break
			}

		}
	}
	return leafValue, nil

}

// ParseArray Parses a given array
func ParseArray(array []interface{}, leaf string) {
	for _, val := range array {
		switch val.(type) {
		case map[string]interface{}:
			FindLeaf(val.(map[string]interface{}), leaf)
		case []interface{}:
			ParseArray(val.([]interface{}), leaf)

		}
	}
}

// GnmiFullPath builds the full path from the prefix and path.
func GnmiFullPath(prefix, path *pb.Path) *pb.Path {
	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElement() != nil {
		fullPath.Element = append(prefix.GetElement(), path.GetElement()...)
	}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}
