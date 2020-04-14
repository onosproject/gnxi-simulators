ARG ONOS_BUILD_VERSION=undefined

FROM onosproject/golang-build:$ONOS_BUILD_VERSION as build

RUN mkdir -p $GOPATH \
    && GO111MODULE=on go get -u \
      github.com/google/gnxi/gnoi_target@6697a080bc2d3287d9614501a3298b3dcfea06df \
      github.com/google/gnxi/gnoi_cert@6697a080bc2d3287d9614501a3298b3dcfea06df 

ENV ONOS_SIMULATORS_ROOT=$GOPATH/src/github.com/onosproject/simulators
ENV GO111MODULE=off
ENV CGO_ENABLED=0

RUN mkdir -p $ONOS_SIMULATORS_ROOT/

COPY cmd/ $GOPATH/src/github.com/onosproject/simulators/cmd/
COPY pkg/ $GOPATH/src/github.com/onosproject/simulators/pkg/

RUN cd $GOPATH/src/github.com/onosproject/simulators && \
    GO111MODULE=on go get github.com/onosproject/simulators/cmd/gnmi_target

FROM alpine:3.11
RUN apk add --update bash openssl curl && rm -rf /var/cache/apk/*
ENV ONOS_SIMULATORS_ROOT=$GOPATH/src/github.com/onosproject/simulators
ENV GNMI_PORT=10161
ENV GNOI_PORT=50001
ENV SIM_MODE=1
ENV HOME=/home/devicesim
RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/*target /usr/local/bin/

COPY configs/target_configs target_configs
COPY tools/scripts scripts
COPY pkg/certs certs

RUN chmod +x ./scripts/run_targets.sh
CMD ["./scripts/run_targets.sh"]
