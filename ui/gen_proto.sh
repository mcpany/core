#!/bin/bash
mkdir -p src/proto
files=$(cd .. && find proto -name "*.proto")
cd ..
npx protoc \
  --plugin=protoc-gen-ts_proto=ui/node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=ui/src/proto \
  --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
  --proto_path=. \
  --proto_path=tmp_proto \
  $files
