export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

ONOS_SIMULATORS_VERSION := latest
ONOS_BUILD_VERSION := stable

all: build images

images: # @HELP build simulators image
images: simulators-docker
	
deps: # @HELP ensure that the required dependencies are in place
	go build -v ./...
	bash -c "diff -u <(echo -n) <(git diff go.mod)"
	bash -c "diff -u <(echo -n) <(git diff go.sum)"

lint: # @HELP run the linters for Go source code
	go list ./... | grep -v /gnmi/modeldat |  xargs -L1 golint -set_exit_status

vet: # @HELP examines Go source code and reports suspicious constructs
	go vet github.com/onosproject/simulators/pkg/... 
	go vet github.com/onosproject/simulators/cmd/...

license_check: # @HELP examine and ensure license headers exist
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR}

gofmt: # @HELP run the go format utility against code in the pkg and cmd directories
	bash -c "diff -u <(echo -n) <(gofmt -d pkg/ cmd/)"

# @HELP build the go binary in the cmd/gnmi_target package
build: test
	go build -o build/_output/gnmi_target ./cmd/gnmi_target

test: deps vet license_check gofmt lint
	go test github.com/onosproject/simulators/pkg/...
	go test github.com/onosproject/simulators/cmd/...

simulators-docker:
	docker build . -f Dockerfile \
	--build-arg ONOS_BUILD_VERSION=${ONOS_BUILD_VERSION} \
	-t onosproject/device-simulator:${ONOS_SIMULATORS_VERSION}

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
