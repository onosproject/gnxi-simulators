<!--
SPDX-FileCopyrightText: 2022 2020-present Open Networking Foundation <info@opennetworking.org>

SPDX-License-Identifier: Apache-2.0
-->

# Simulators

[![Build Status](https://api.travis-ci.org/onosproject/gnxi-simulators.svg?branch=master)](https://travis-ci.org/onosproject/gnxi-simulators)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/onosproject/simulators?status.svg)](https://godoc.org/github.com/onosproject/simulators)

Simple simulators, used for integration testing of ONOS interactions with devices and various orchestration entities, e.g:

- Configuring devices via gNMI and OpenConfig
- Controlling operation of devices via gNOI
- Shaping pipelines and controlling traffic flow via P4 programs and P4Runtime

The simulator facilities are available as Go package libraries, executable commands and as published Docker containers.

# Additional Documentation

[How to run](docs/README.md) device simulator and related commands.
