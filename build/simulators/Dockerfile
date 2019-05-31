ARG ONOS_BUILD_VERSION=stable

FROM onosproject/onos-config-build:$ONOS_BUILD_VERSION as builder
COPY . /go/src/github.com/onosproject/simulators
RUN cd /go/src/github.com/onosproject/simulators && make build

FROM alpine:3.8

USER nobody

COPY --from=builder /go/src/github.com/onosproject/simulators/build/_output/simulators /usr/local/bin/gnmi_target

ENTRYPOINT ["gnmi_target"]

