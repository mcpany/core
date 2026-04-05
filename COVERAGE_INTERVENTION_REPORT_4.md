# Coverage Intervention Report

* **Target:** `server/pkg/lifecycle/reaper.go` (`RegisterSubagent`, `RecordHeartbeat`, `PruneIntent`, `GetLeaseStatus`)
* **Risk Profile:** This code deals with resource leases, subagent lifecycles, sweeping, and process isolation boundary cleanup. These functions had critical logic loops and 0% coverage on functions like `RegisterSubagent`. It fits the 'High Risk' profile since failures in resource tracking or state cleanup can lead to zombie processes, resource exhaustion, and security boundary violations across tenants.
* **New Coverage:**
  *   **`RegisterSubagent`:** Guarded the happy path (appending subagent ID) and edge cases (rejecting if lease not found or not `ACTIVE`).
  *   **`RecordHeartbeat`:** Guarded missing intent errors and failure if the lease is inactive.
  *   **`PruneIntent` & `GetLeaseStatus`:** Added edge case testing for when the intent is not found.
  *   Coverage for `server/pkg/lifecycle/reaper.go` improved from ~73% to ~96.8%.
* **Verification:** Confirmed that `go test ./pkg/lifecycle/...` runs green and that all tests passed without side effects.
