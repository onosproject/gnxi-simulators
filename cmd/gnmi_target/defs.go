// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"time"

	"github.com/onosproject/gnxi-simulators/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

var (
	bindAddr            = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	configFile          = flag.String("config", "", "IETF JSON file for target startup config")
	readOnlyPath        = `elem:<name:"system" > elem:<name:"openflow" > elem:<name:"controllers" > elem:<name:"controller" key:<key:"name" value:"main" > > elem:<name:"connections" > elem:<name:"connection" key:<key:"aux-id" value:"0" > > elem:<name:"state" > elem:<name:"address" > `
	randomEventInterval = time.Duration(5) * time.Second
)

type server struct {
	*gnmi.Server
	Model               *gnmi.Model
	configStruct        ygot.ValidatedGoStruct
	UpdateChann         chan *pb.Update
	readOnlyUpdateValue *pb.Update
}

type streamClient struct {
	target  string
	sr      *pb.SubscribeRequest
	stream  pb.GNMI_SubscribeServer
	errChan chan<- error
}
