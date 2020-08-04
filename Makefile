export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

ONOS_SIMULATORS_VERSION := latest
ONOS_BUILD_VERSION := v0.6.0

all: build images

images: # @HELP build simulators image
images: simulators-docker


deps: # @HELP ensure that the required dependencies are in place
	go build -v ./...
	bash -c "diff -u <(echo -n) <(git diff go.mod)"
	bash -c "diff -u <(echo -n) <(git diff go.sum)"

linters: # @HELP examines Go source code and reports coding problems
	golangci-lint run

license_check: # @HELP examine and ensure license headers exist
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR}


# @HELP build the go binary in the cmd/gnmi_target package
build:
	go build -o build/_output/gnmi_target ./cmd/gnmi_target

test: build deps license_check linters
	go test github.com/onosproject/simulators/pkg/...
	go test github.com/onosproject/simulators/cmd/...

simulators-docker:
	docker build . -f Dockerfile \
	--build-arg ONOS_BUILD_VERSION=${ONOS_BUILD_VERSION} \
	-t onosproject/device-simulator:${ONOS_SIMULATORS_VERSION}

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image onosproject/device-simulator:${ONOS_SIMULATORS_VERSION}

publish: # @HELP publish version on github and dockerhub
	./../build-tools/publish-version ${VERSION} onosproject/device-simulator

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output
	rm -rf ./vendor
	rm -rf ./cmd/gnmi_target/gnmi_target

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
