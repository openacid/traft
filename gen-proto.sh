#!/bin/sh

# protoc --proto_path=. --go_out=plugins=grpc:. traft.proto
# protoc --proto_path=. --gofast_out=plugins=grpc:. traft.proto


# go get github.com/gogo/protobuf/protoc-gen-gogofast
# go get github.com/gogo/protobuf/protoc-gen-gogofaster
# go get github.com/gogo/protobuf/protoc-gen-gogoslick

protoc -I=. \
    -I=$GOPATH/src \
    -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
    --gogofaster_out=plugins=grpc:. \
    traft.proto
