#!/bin/sh

protoc --proto_path=. --go_out=plugins=grpc:. traft.proto
