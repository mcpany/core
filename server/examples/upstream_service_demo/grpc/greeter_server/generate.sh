#!/bin/bash

export PATH
PATH=$(pwd)/../../../../build/env/bin:$PATH
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/greeter.proto
