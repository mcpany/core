#!/bin/bash
sed -i 's/fmt.Sprintf("%s=%s", key, val)/key + "=" + val/g' server/pkg/tool/types.go
sed -i 's/fmt.Sprintf("%s=%s", k, v)/k + "=" + v/g' server/pkg/tool/types.go
sed -i 's/fmt.Sprintf("%s=%s", name, secretValue)/name + "=" + secretValue/g' server/pkg/tool/types.go
sed -i 's/fmt.Sprintf("%s=%s", name, valStr)/name + "=" + valStr/g' server/pkg/tool/types.go

sed -i 's/fmt.Sprintf("%s=%s", k, v)/k + "=" + v/g' server/pkg/upstream/mcp/bundle.go

sed -i 's/fmt.Sprintf("%s=%s", k, v)/k + "=" + v/g' server/pkg/upstream/mcp/docker_transport.go

sed -i 's/fmt.Sprintf("%s=%s", key, val)/key + "=" + val/g' server/pkg/upstream/mcp/streamable_http.go
sed -i 's/fmt.Sprintf("%s=%s", k, v)/k + "=" + v/g' server/pkg/upstream/mcp/streamable_http.go
