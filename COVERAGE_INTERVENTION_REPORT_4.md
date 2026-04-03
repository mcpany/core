# Coverage Intervention Report

**Target:** `server/pkg/lifecycle/reaper.go`

**Risk Profile:**
The `SubagentReaper` component is crucial for governing resource limits and lifecycles (via leases) associated with dynamically executed subagents. Uncaught failures or incorrect lease handling could lead to orphaned agent processes, unreleased resources (zombies), or incorrect execution states. Despite its importance in avoiding resource leaks, the file was identified with a moderate-to-low test coverage (69.8%), specifically lacking test cases for the central `RegisterSubagent` function and multiple edge-case validation checks (e.g. attempting to interact with missing or inactive leases).

**New Coverage:**
Comprehensive testing has been introduced via modifications to `server/pkg/lifecycle/reaper_test.go`, elevating the coverage of `reaper.go` to exactly 100.0%. The following specific logic paths are now rigorously guarded:
*   `RegisterSubagent`: Adding a subagent to a valid lease, handling unregistered intent strings, and preventing registration to inactive/pruned leases.
*   `RecordHeartbeat`: Properly managing missing intents and blocking heartbeats for non-active leases.
*   `PruneIntent`: Emitting correct error states when trying to manually prune a nonexistent intent.
*   `GetLeaseStatus`: Fetching status for invalid intent IDs.
*   `Start` / `Stop` Synchronization: Ensuring the background cleanup goroutines exit properly via context cancellation (`ctx.Done()`) or via explicit shutdown signal (`r.quit` via `Stop()`).

**Verification:**
*   Confirmed that `make test` and `make lint` passed cleanly across the repository.
*   The `go test -coverprofile` check on `./pkg/lifecycle/...` confirms statements in `reaper.go` are entirely verified.
