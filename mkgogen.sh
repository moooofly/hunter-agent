#!/usr/bin/env bash

# To generate:
#
# cd $(go env GOPATH)/src/moooofly/hunter-agent
# ./mkgogen.sh

OUTDIR="$(go env GOPATH)/src"

protoc -I=. --go_out=plugins=grpc:$OUTDIR proto/dump.proto
