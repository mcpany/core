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

---

* **Target:** `server/pkg/lifecycle/reaper.go`
* **Risk Profile:** The lifecycle reaper is a critical component for managing background subagent sessions, keeping track of active intents, and cleaning up memory/connections to avoid process leaks or zombie tasks. High risk if there are unregistered/rogue subagents or panic inducing edge cases when dealing with them. It also had only 73% coverage overall with 0% coverage on important functions like `RegisterSubagent`.
* **New Coverage:** Added coverage for `RegisterSubagent` happy and error paths (when lease doesn't exist or is pruned), as well as covering error paths for `RecordHeartbeat`, `PruneIntent`, and `GetLeaseStatus` when the lease is missing or already pruned. Coverage in the package was raised to 96.8%.
* **Verification:** Confirmed that `go test` in the package passed cleanly, as well as `make test` for the whole repo.
