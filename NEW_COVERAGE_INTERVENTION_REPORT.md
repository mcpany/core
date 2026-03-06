# Coverage Intervention Report

**Top 10 High-Risk Untested Components Discovered:**
1. `server/pkg/middleware/global_ratelimit.go` (Central entry point configuration/Redis)
2. `server/pkg/upstream/filesystem/provider/sftp.go` (Untested file I/O operations)
3. `server/pkg/upstream/vector/pinecone.go` (Pinecone queries lacking coverage)
4. `server/pkg/upstream/mcp/streamable_http.go` (Healthchecks and connections are untested)
5. `server/pkg/upstream/mcp/docker_transport.go` (Read/Write behaviors have low coverage)
6. `server/pkg/tool/types.go` (Execute functions for MCP Tool wrapping)
7. `server/pkg/middleware/audit.go` (Subscriptions and history tracking)
8. `server/pkg/middleware/semantic_cache_postgres.go` (Closing/opening connection pools)
9. `server/pkg/app/server.go` (Run and reconcile loops)
10. `server/pkg/upstream/grpc/grpc_pool.go` (Closing operations on the connection pools)

**Target:** `server/pkg/upstream/mcp/streamable_http.go` and `server/pkg/upstream/mcp/bundle_local_transport.go`

**Risk Profile:**
These components manage upstream MCP health checks, tool/prompt/resource definitions, and local bundle connections via exec. Both had segments with 0% coverage on critical initialization or reporting pathways, which poses a severe risk in correctly routing and tracking upstream availability or connection state.

**New Coverage:**
* `streamable_http.go`: Covered `CheckHealth` (both no checker and status reporting code paths) and `Definition` for translating prompt arguments into expected MCP property schemas (100% coverage on previously untested functions).
* `bundle_local_transport.go`: Covered `Connect`, ensuring we generate correct `StdioTransport` components when running local commands (100% coverage on `Connect`).

**Verification:**
Verified that `go test ./pkg/upstream/mcp/...` passes cleanly without regression. Coverage in the mcp upstream module increased from 81.7% to 84.5%. Linting passes cleanly.
