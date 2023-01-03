#!/usr/bin/env bash
cd ..
## protobuf
protoc --proto_path=api/v1 --proto_path=api/v1/third_party --go_out=plugins=grpc,paths=source_relative:service/api user_service.proto
