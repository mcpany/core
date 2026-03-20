#!/bin/bash

# We just need to remove the go_features import and usage temporarily
# to generate the TS definitions, as the TS definitions don't care about Go features.

find proto -type f -name "*.proto" -exec sed -i 's/import "google\/protobuf\/go_features.proto";//g' {} +
find proto -type f -name "*.proto" -exec sed -i 's/option features.(pb.go).api_level = API_OPAQUE;//g' {} +

export PATH="$PWD/protoc3/bin:$PATH"

export PATH="/app/ui/node_modules/.bin:$PATH"

cd proto
for dir in api/v1 config/v1 mcp_options/v1 mcp_router/v1 topology/v1 examples/weather/v1 admin/v1 bus; do
    echo "Generating protos in $dir..."
    protoc \
      --plugin=protoc-gen-ts_proto=/app/ui/node_modules/.bin/protoc-gen-ts_proto \
      --ts_proto_out=. \
      --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
      --proto_path=.. \
      --proto_path=/app/protoc3/include \
      --experimental_editions \
      ../proto/$dir/*.proto
done

# restore
git checkout -- .
