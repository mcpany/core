# Coverage Intervention Report

**Target:** `server/pkg/storage/postgres/store.go` and `server/pkg/storage/sqlite/store.go`

**Risk Profile:**
These files handle critical interactions with the database layer, specifically persisting global state and logging interactions. We identified that the `SaveLog` method, responsible for saving system and session activity logs to the database, was entirely untested. Missing tests in database components pose a high risk of runtime failures and data loss if regressions are introduced.

**New Coverage:**
- `server/pkg/storage/postgres/store_test.go`: Added specific tests for `SaveLog` to verify correct query execution and error handling using `sqlmock`.
- `server/pkg/storage/sqlite/store_test.go`: Added specific tests for `SaveLog` leveraging the real temporary database approach used in the rest of the SQLite suite to confirm that a valid log entry is safely committed to the database.

**Verification:**
- Executed `go test ./pkg/storage/... -v` to confirm new tests pass cleanly.
- Verified test coverage increased for both storage layers.
- Ran `make test` and `make lint` from the project root. `make lint` passed with 100% compliance.
