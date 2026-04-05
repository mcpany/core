# Coverage Intervention Report

* **Target:** `server/pkg/lifecycle/reaper.go`
* **Risk Profile:** The `SubagentReaper` manages intent-bound leases, subagent sessions, and resource lifecycles. This is a core component that guarantees resources are appropriately pruned and active branches are tracked without running into zombie processes. Prior to this intervention, several edge cases around missing intents and inactive leases were not tested, leaving these critical logic paths vulnerable to silent failures or panics if assumptions broke in the future.
* **New Coverage:**
  - `RegisterSubagent`: Now tested for missing intent and attempting to register to an inactive lease.
  - `RecordHeartbeat`: Now tested for missing intent and extending a heartbeat on an inactive lease.
  - `PruneIntent`: Now tested for missing intent.
  - `GetLeaseStatus`: Now tested for missing intent.
  - `Start`: Context cancellation flow is verified to ensure the daemon goroutine cleanly halts.
  These cases raised the test coverage of `reaper.go` from 73% to 100%.
* **Verification:** Clean `make test` and `make docker-lint` (or equivalent go commands) with no regressions. The new tests pass seamlessly as part of the `lifecycle` package suite.