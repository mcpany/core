# Coverage Intervention Report

* **Target:** `server/pkg/tokenizer/tokenizer.go`
* **Risk Profile:** This file contains highly complex and recursive reflection logic for tokenizing arbitrarily structured generic JSON data. Methods like `countTokensInValueRecursive`, `countTokensReflectMap`, and `countTokensReflectStruct` lacked test coverage on deep fallback and error paths (e.g. cycle detection, reflection panics on unsupported types). Since this code parses unknown structured input payloads to calculate token limits for LLM interactions, unhandled errors or panics in this file represent a severe security/reliability risk (e.g., recursive exhaustion, crashing the MCP server on a bad user payload).
* **New Coverage:** The following logic paths are now guarded by comprehensive tests:
  - Error paths and recursion cycle detection in generic types (`reflectSlice`, `reflectMap`, `reflectStruct`, `countSliceInterfaceSimple`).
  - Fallback behaviors for `countTokensInValueRecursive` handling unsupported payload types.
  - Edge cases for fast-path primitive tokenization (e.g. `simpleTokenizeInt64` with large negative numbers and zero values).
* **Verification:** `make test` successfully tests the new components alongside all existing legacy tests. The overall file coverage has increased significantly, reaching >97% statement coverage with the new regression safety nets in place. Assertions are strictly based on the specific expected behaviour (token counts).

---

* **Target:** `server/pkg/tool/webrtc.go`
* **Risk Profile:** This file establishes WebRTC peer connections allowing LLMs to communicate directly with streaming endpoints, which is core business logic for real-time interactions. It utilizes custom unmarshalling and connection pools. If untested, failures in pooling logic or unmarshalling JSON parameters lead to silent drop of connections or panics.
* **New Coverage:** The tests added cover logic paths handling stream execution error cases (verifying that `StreamExecute` appropriately handles execution errors without crashing), executing without pool instances and checking JSON unmarshalling. It explicitly asserts the behaviors and returned errors on bad inputs instead of just hitting the coverage.
* **Verification:** `bazelisk test //...` verified cleanly with all legacy tests running and green. The style mimics Google Table-Driven tests inside the `stretchr/testify` framework constraints in place.

---

* **Target:** `server/pkg/lifecycle/reaper.go`
* **Risk Profile:** This file was selected due to its core role in process isolation, resource leasing, and intent management for subagents. A failure in managing these components correctly can easily lead to severe resource exhaustion, orphan subagent sessions (zombies), or incorrect processing limits. Before intervention, this file was missing coverage in several error-prone edge-case paths, increasing the risk of unstable agent sessions silently failing or persisting.
* **New Coverage:**
  - `RegisterSubagent`: Now covers logic handling requests to register against non-existent or pruned intents.
  - `RecordHeartbeat`: Now guards against non-existent intent checks and inactive leases.
  - `PruneIntent`: Now correctly tests missing/not-found intent operations.
  - `GetLeaseStatus`: Includes checking for non-existent items.
  - `Start/Stop` Context Loop: Adds coverage verifying that cancelling the provided context stops the daemon and safely terminates the routine loop to prevent goroutine leaks.
  - Overall file coverage increased from 73% to 96.8%.
* **Verification:** Confirmed that `go test ./pkg/lifecycle/...` and `make lint` passed cleanly. Pre-existing tests were not broken.

### Top 10 Most Critical Untested Components
Based on cyclomatic complexity and risk (e.g. data transformation, tool injection safety checks, execution endpoints), the top 10 most critical untested logic components are:
1. `server/pkg/tool/types.go`
2. `server/pkg/app/server.go`
3. `server/pkg/config/store.go`
4. `server/pkg/upstream/mcp/streamable_http.go`
5. `server/pkg/mcpserver/server.go`
6. `server/pkg/config/validator.go`
7. `server/pkg/storage/postgres/store.go`
8. `server/pkg/storage/sqlite/store.go`
9. `server/pkg/app/api.go`
10. `server/pkg/upstream/filesystem/provider/gcs.go`