#!/bin/bash
sed -i 's/"google.golang.org\/protobuf\/types\/known\/structpb"//' server/pkg/upstream/http/http_input_schema_bug_test.go
sed -i 's/mcpsdk "github.com\/modelcontextprotocol\/go-sdk\/mcp"//' server/tests/integration/upstream/webhooks_e2e_test.go
