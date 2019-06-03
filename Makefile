export CGO_ENABLED=0

.PHONY: build

ONOS_SIMULATORS_VERSION := latest
ONOS_BUILD_VERSION := stable

all: image

image: # @HELP build simulators image
	docker run -it -v `pwd`:/go/src/github.com/onosproject/simulators -w "/go/src/github.com/onosproject/simulators"  onosproject/onos-config-build:${ONOS_BUILD_VERSION} build -e GO111MODULE=on
	docker build . -f build/simulators/Dockerfile \
	--build-arg ONOS_BUILD_VERSION=${ONOS_BUILD_VERSION} \
	-t onosproject/simulators:${ONOS_SIMULATORS_VERSION}

deps: # @HELP ensure that the required dependencies are in place
	go build -v ./...

lint: # @HELP run the linters for Go source code
	go list ./... | grep -v /gnmi/modeldat |  xargs -L1 golint -set_exit_status

vet: # @HELP examines Go source code and reports suspicious constructs
	go vet github.com/onosproject/simulators/pkg/... 
	go vet github.com/onosproject/simulators/cmd/...

license_check: # @HELP examine and ensure license headers exist
	./build/licensing/boilerplate.py -v

gofmt: # @HELP run the go format utility against code in the pkg and cmd directories
	bash -c "diff -u <(echo -n) <(gofmt -d pkg/ cmd/)"

# @HELP build the go binary in the cmd/gnmi_target package
build: test
	export GOOS=linux
	export GOARCH=amd64
	go build -o build/_output/gnmi_target ./cmd/gnmi_target

test: deps vet license_check gofmt lint
	go test github.com/onosproject/simulators/pkg/...
	go test github.com/onosproject/simulators/cmd/...


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
