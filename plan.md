1. **Analyze**
   - Goal: Identify "Dark Matter" in `server/pkg/storage/sqlite/store.go` and `server/pkg/storage/postgres/store.go`, focusing on untested methods.
   - Using the generated `sqlite_test/coverage.dat`, the method `GetRecentLogs` (lines 876-921) has zero coverage.
   - Other untested methods like `SaveLog` (lines 845-874), `ListCredentials` (1327-1366), `GetCredential` (1368-1401) also exist.
   - I will focus on `GetRecentLogs` as it is an essential piece for UI log broadcasters and contains DB scan complexity with `sql.NullString` and JSON unmarshaling logic. It's high risk because a schema mismatch or unmarshaling error could break the log viewer. I will also write tests for `SaveLog`.

2. **Test Implementation**
   - I will modify `server/pkg/storage/sqlite/store_test.go` and `server/pkg/storage/postgres/store_test.go`.
   - I will create `TestStore_Logs` in both files using the Table-Driven Test pattern (or subtests) commonly found in Go codebases.
   - The test will insert some logs (`SaveLog`) and then read them (`GetRecentLogs`) and check if the order and limits are respected.
   - Mocking: the tests seem to use an actual in-memory SQLite and temporary Postgres instances (via test helpers). I will follow the established `setupTestDB` pattern.

3. **The Regression Gate**
   - After implementing the test, I will run `go test` and `bazelisk test` to ensure it passes.
   - I will fix any failures without breaking legacy tests.

4. **Impact Report**
   - I will update `COVERAGE_INTERVENTION_REPORT.md` (or create a new one like `COVERAGE_INTERVENTION_REPORT_4.md`) with the required details.
