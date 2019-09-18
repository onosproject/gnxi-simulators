ARG ONOS_BUILD_VERSION=stable
FROM golang:1.11-alpine AS build
LABEL maintainer="Sean Condon <sean@opennetworking.org>, Adib Rastegarnia <adib@opennetworking.org> "
LABEL description="Builds a gNMI/gNOI simulator on a Debian distribution"

RUN apk add --update bash openssl curl && rm -rf /var/cache/apk/*

RUN apk update \
    && apk add \
      ca-certificates \
    && apk update

RUN apk add \
    git \
    iputils \
    net-tools \
    psmisc \
    procps \
    sudo


RUN mkdir -p $GOPATH \
    && go get -u \
      github.com/google/gnxi/gnmi_capabilities \
      github.com/google/gnxi/gnmi_get \
      github.com/google/gnxi/gnmi_set \
      github.com/openconfig/gnmi/cmd/gnmi_cli \
      github.com/google/gnxi/gnoi_target \ 
      github.com/google/gnxi/gnoi_cert 
      

RUN go install -v \
      github.com/google/gnxi/gnmi_capabilities \
      github.com/google/gnxi/gnmi_get \
      github.com/google/gnxi/gnmi_set \
      github.com/openconfig/gnmi/cmd/gnmi_cli \
      github.com/google/gnxi/gnoi_target \ 
      github.com/google/gnxi/gnoi_cert


ENV ONOS_SIMULATORS_ROOT=$GOPATH/src/github.com/onosproject/simulators
ENV GO111MODULE=off
ENV CGO_ENABLED=0

RUN mkdir -p $ONOS_SIMULATORS_ROOT/

COPY cmd/ $GOPATH/src/github.com/onosproject/simulators/cmd/
COPY pkg/ $GOPATH/src/github.com/onosproject/simulators/pkg/


RUN cd $GOPATH/src/github.com/onosproject/simulators/cmd/gnmi_target && go install

FROM alpine:3.9
RUN apk add --update bash openssl curl && rm -rf /var/cache/apk/*
ENV ONOS_SIMULATORS_ROOT=$GOPATH/src/github.com/onosproject/simulators
ENV GNMI_PORT=10161
ENV GNOI_PORT=50001
ENV SIM_MODE=1
ENV HOME=/home/devicesim
RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/ /usr/local/bin

COPY configs/target_configs target_configs
COPY tools/scripts scripts
COPY pkg/certs certs

RUN chmod +x ./scripts/run_targets.sh

CMD ["./scripts/run_targets.sh"]
