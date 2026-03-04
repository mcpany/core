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

**Target:** `server/pkg/middleware/global_ratelimit.go`

**Risk Profile:**
This component manages global rate limiting for incoming MCP requests. It relies on caching, handles context propagation for different user identity signals (API Key, User ID, HTTP Headers), and connects dynamically to Redis. Prior to intervention, key methods handling configuration updates, caching layer integration, partitioned key resolution, and configuration hash calculations had zero or extremely low test coverage. Because this is the primary edge defense for preventing DoS/abuse on upstream tools, any untested regressions in routing logic could lead to service disruption or incorrect limit enforcement.

**New Coverage:**
*   `UpdateConfig`: Added coverage to ensure that changing rate limiting properties at runtime propagates to the middleware state correctly (100% coverage).
*   `getPartitionKey`: Added robust table-driven tests evaluating the context extraction logic. It now verifies correct resolution of IP, User ID, and API keys sourced both via typical Context patterns and directly off incoming HTTP Request headers (100% coverage).
*   `calculateConfigHash`: Confirmed that structurally identical vs distinct Redis configuration sets are mapped properly to avoid colliding cached client pools (100% coverage).
*   `getRedisClient`: Validated the cache lookups for Redis connections to assure we are returning active cached clients on identical configs, preventing leakages from excessive redundant client spawning (81.8% coverage).
*   `getLimiter`: Tested retrieval, fallback logic without Redis configs, and caching behaviors, ensuring we accurately clamp invalid burst inputs (83.3% coverage).

**Verification:**
Confirmed that `go test` specific to the middleware package passes cleanly and coverage has drastically improved (>86%). Linting is passing.
