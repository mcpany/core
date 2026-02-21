# Coverage Intervention Report

* **Target:** `server/pkg/resource/resource.go` (Manager implementation)
* **Risk Profile:** The `Manager` struct uses double-checked locking for caching resources and manages concurrent access to the resource map. It also handles callbacks for list changes. This is a classic high-risk area for concurrency bugs (race conditions, deadlocks, inconsistent state). The existing test coverage was purely single-threaded, leaving this complex logic effectively untested under load.
* **New Coverage:**
    *   **Concurrency Verification:** Added `server/pkg/resource/resource_concurrency_test.go` which introduces three new test scenarios:
        1.  `TestManager_ConcurrentAccess`: Spawns 50 concurrent goroutines performing mix of `AddResource`, `RemoveResource`, and `ListResources` to stress-test the locking mechanisms.
        2.  `TestManager_CallbackConcurrency`: Verifies that `OnListChanged` callbacks are executed correctly under concurrent load without causing deadlocks.
        3.  `TestManager_ClearResourcesConcurrency`: Tests the interaction between `ClearResourcesForService` and other operations to ensure atomic clearing and consistency.
    *   **Thread Safety:** Validated that the `Manager` is thread-safe using `go test -race`.
* **Verification:**
    *   `go test -v -race ./server/pkg/resource/...` passed cleanly.
    *   `make test` (full suite) was run and core packages passed (failures in upstream/http and upstream/mcp were identified as environment timeouts unrelated to changes).
    *   `make lint` shows no new issues in the touched package.
