# Coverage Intervention Report

**Target:** `server/pkg/storage/postgres/store.go` - `(*Store).Load()`

**Risk Profile:**
This code was selected because the `Load` method runs 5 parallel database queries via `errgroup.Group` to construct the entire application configuration from Postgres. It had a high cyclomatic complexity (27) due to multiple JSON unmarshalling and scanning steps, combined with extremely low/zero test coverage for the concurrent paths and error states. This is core data transformation logic that forms the backbone of configuration loading on startup. A failure here brings down the router.

**New Coverage:**
- **Happy Path:** Validates concurrent fetching of Upstream Services, Users, Global Settings, Collections, and Profiles using table-driven tests and `go-sqlmock`.
- **Query Error States:** Validates that query execution failures on individual tables immediately bubble up an error.
- **Collection Ignore State:** Validates that `service_collections` is the only query where errors are logged rather than failing the entire load operation.
- **Data Transformation Failures:** Validates that database parsing failures (`Scan` errors) and JSON unmarshalling failures are correctly intercepted and reported.

**Verification:**
Verified that all parallel paths remain hermetic by dynamically creating fresh DB mocks on each test iteration and using `mock.MatchExpectationsInOrder(false)`.
`bazel test //server/...` passed cleanly.
`bazel run //:lint` passed cleanly.
