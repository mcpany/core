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

**Target:** `server/pkg/middleware/audit.go` & `server/pkg/middleware/semantic_cache_postgres.go`

**Risk Profile:**
*   `server/pkg/middleware/audit.go`: The Audit middleware logs all tools and executions to various storage sinks. It has broadcasting and unsubscription logic, which handles real-time metrics and streaming traces. Missing test coverage for subscription mechanisms (`SubscribeWithHistory`, `GetHistory`, `Unsubscribe`, and `Write`) presents a risk of concurrent map failures, history leakage, and dropped events during runtime operations.
*   `server/pkg/middleware/semantic_cache_postgres.go`: This coordinates caching and returning expensive AI vector search results. `NewPostgresVectorStore` and `Close` were missing coverage, posing a risk of runtime panics, resource leakage, and incorrect dependency bootstrapping when initialized with valid or invalid connections.

**New Coverage:**
*   `SubscribeWithHistory` / `GetHistory` / `Unsubscribe`: Implemented full coverage mimicking the test suite behavior. Ensures history channels are accurately appended and effectively drained, validating atomicity over unsubscribed channel handling (100% coverage).
*   `Write`: Validated that direct event emissions route appropriately to initialized storage drivers without bypassing redaction bounds (100% coverage).
*   `NewPostgresVectorStore`: Introduced tests asserting connection initialization errors like `EmptyDSN` and `InvalidDSN` are adequately trapped and raised securely (100% coverage).
*   `Close`: Confirmed connections are appropriately released back to the pool to resolve underlying socket leakage (100% coverage).

**Verification:**
Confirmed that `go test` executed cleanly, and the `server/pkg/middleware` test suite saw significant coverage increase (reaching 100% coverage on the patched methods). Linting successfully passes.
