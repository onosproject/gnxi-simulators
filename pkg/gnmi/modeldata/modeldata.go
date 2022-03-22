// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package modeldata contains the following model data in gnmi proto struct:
//	openconfig-interfaces 2.0.0,
//	openconfig-openflow 0.1.0,
//	openconfig-platform 0.5.0,
//	openconfig-system 0.2.0.
package modeldata

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	// OpenconfigInterfacesModel is the openconfig YANG model for interfaces.
	OpenconfigInterfacesModel = "openconfig-interfaces"
	// OpenconfigOpenflowModel is the openconfig YANG model for openflow.
	OpenconfigOpenflowModel = "openconfig-openflow"
	// OpenconfigPlatformModel is the openconfig YANG model for platform.
	OpenconfigPlatformModel = "openconfig-platform"
	// OpenconfigSystemModel is the openconfig YANG model for system.
	OpenconfigSystemModel = "openconfig-system"
)

var (
	// ModelData is a list of supported models.
	ModelData = []*pb.ModelData{{
		Name:         OpenconfigInterfacesModel,
		Organization: "OpenConfig working group",
		Version:      "2017-07-14",
	}, {
		Name:         OpenconfigOpenflowModel,
		Organization: "OpenConfig working group",
		Version:      "2017-06-01",
	}, {
		Name:         OpenconfigPlatformModel,
		Organization: "OpenConfig working group",
		Version:      "2016-12-22",
	}, {
		Name:         OpenconfigSystemModel,
		Organization: "OpenConfig working group",
		Version:      "2017-07-06",
	}}
)
