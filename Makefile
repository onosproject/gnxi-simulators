# SPDX-FileCopyrightText: 2022 2020-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

ONOS_SIMULATORS_VERSION := latest
ONOS_BUILD_VERSION := v0.6.0

all: build images

build-tools:=$(shell if [ ! -d "./build/build-tools" ]; then cd build && git clone https://github.com/onosproject/build-tools.git; fi)
include ./build/build-tools/make/onf-common.mk

images: # @HELP build simulators image
images: simulators-docker

# @HELP build the go binary in the cmd/gnmi_target package
build: deps
	go build -o build/_output/gnmi_target ./cmd/gnmi_target

test: build deps license linters
	go test github.com/onosproject/gnxi-simulators/pkg/...
	go test github.com/onosproject/gnxi-simulators/cmd/...

jenkins-test:  # @HELP run the unit tests and source code validation producing a junit style report for Jenkins
jenkins-test: deps license linters
	TEST_PACKAGES=github.com/onosproject/gnxi-simulators/... ./build/build-tools/build/jenkins/make-unit

simulators-docker:
	docker build . -f Dockerfile \
	--build-arg ONOS_BUILD_VERSION=${ONOS_BUILD_VERSION} \
	-t onosproject/device-simulator:${ONOS_SIMULATORS_VERSION}

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image onosproject/device-simulator:${ONOS_SIMULATORS_VERSION}

publish: # @HELP publish version on github and dockerhub
	./build/build-tools/publish-version ${VERSION} onosproject/device-simulator

jenkins-publish: # @HELP Jenkins calls this to publish artifacts
	./build/bin/push-images
	./build/build-tools/release-merge-commit

clean:: # @HELP remove all the build artifacts
	rm -rf ./build/_output
	rm -rf ./vendor
	rm -rf ./cmd/gnmi_target/gnmi_target

