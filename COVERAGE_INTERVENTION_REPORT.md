# Coverage Intervention Report

**Target:** `server/pkg/storage/postgres/store.go` (specifically `(*Store).Load`)

**Risk Profile:**
The `Load` function in the PostgreSQL storage adapter is responsible for bootstrapping the entire `McpAnyServerConfig` object on startup, which includes services, users, settings, collections, and profiles. With a cyclomatic complexity of 27, it employs multiple concurrent goroutines using a wait group to query five different tables, deserialize their JSON configurations via protojson, and aggregate the results. Prior to this intervention, the `Load` logic was virtually untested directly via unit tests. Failures in this function, such as bad configuration syntax or database disruption, could lead to startup panics or silent initialization regressions without warning. Given that it loads authentication data (users) and core capabilities (services), it was a critical, high-risk gap in coverage.

**New Coverage:**
The newly introduced tests in `server/pkg/storage/postgres/store_load_test.go` guard several critical logic paths:
- **Happy Path:** Validates that the function accurately queries, parses, and aggregates data from all 5 distinct tables (`upstream_services`, `users`, `settings`, `service_collections`, and `profiles`) concurrently, proving the data structure resolves successfully into a `configv1.McpAnyServerConfig` object.
- **Query Error Path:** Asserts that an error occurring during the SQL query phase (e.g., failure querying `upstream_services`) correctly returns a wrapped `err` rather than panicking or swallowing the error in a detached goroutine.
- **Scan/Unmarshal Error Path:** Tests resilience against corrupted JSON stored in the database (e.g., malformed service config), ensuring the parsing failure properly bubbles up and short-circuits the loading process safely.
- **Optional Data Error Path (Settings Not Found):** Verifies that missing row data specifically for `global` settings handles `sql.ErrNoRows` gracefully, avoiding an error out and correctly passing nil settings back.

**Verification:**
The new test suite was fully run against the local codebase.
- `bazelisk test //server/pkg/storage/postgres/...` passed completely.
- `bazelisk coverage //server/pkg/storage/postgres/...` generated proper coverage results showing increased coverage for `store.go`.
