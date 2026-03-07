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

**Target:** `server/pkg/upstream/mcp/streamable_http.go`

**Risk Profile:**
This component manages connection health checking and shutdown routines for `streamable_http` MCP tool upstreams. Prior to intervention, `CheckHealth` and `Shutdown` lacked coverage. Because this deals with managing service liveliness (status up/down states) and performing lifecycle cleanup (like properly removing large temporary bundle directories off the disk when services unregister), any regressions could lead to zombie connections, failure to register/unregister MCP resources safely, and unchecked disk leakage from unremoved directories.

**New Coverage:**
*   `CheckHealth`: Added coverage checking the proxy to `checker.Check()`. Verified scenarios for when the checker is absent, when it reports `StatusUp`, and when it reports `StatusDown` with an error message. (100% coverage achieved).
*   `Shutdown`: Added coverage testing cleanup loops and graceful stops. Verified `checker.Stop()` invocation and asserted the successful teardown of the temporary `bundle_dir` using `os.RemoveAll()` when specific `serviceIDs` untrack themselves. (Achieved 93.3% coverage).

**Verification:**
Confirmed that `go test` specific to the streamable package passed cleanly, test coverage is high, and the global tests pass via `make test-fast`. Linting via `make lint` is passing cleanly.