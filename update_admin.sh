#!/bin/bash
unzip -o protoc-29.3-linux-x86_64.zip -d protoc
export PATH="$PATH:$(pwd)/protoc/bin:$(go env GOPATH)/bin"
cd proto
for f in $(find . -name "*.proto"); do protoc --go_out=../server --go_opt=paths=source_relative --go-grpc_out=../server --go-grpc_opt=paths=source_relative -I=. -I=../protoc/include $f; done
cd ..
bazelisk test //server/pkg/admin/...
