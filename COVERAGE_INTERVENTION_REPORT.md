# Coverage Intervention Report

**Target:** `server/pkg/app/server.go` (`reconcileServices`)
**Risk Profile:** High cyclomatic complexity (57) combined with zero test coverage. The `reconcileServices` function is central to the MCP Any server's lifecycle, managing the addition, modification, disablement, and removal of upstream service configurations dynamically. It acts as a core data transformation and logic gateway.
**New Coverage:**
The new test (`server/pkg/app/server_reconcile_test.go`) guards the following logic paths using table-driven tests and hermetic mocks:
- Adding entirely new services to the registry.
- Ignoring existing services that have not changed.
- Updating configurations for existing services.
- Removing services that are missing from the updated configuration.
- Disabling active services via the `Disable` flag (which acts as an explicit removal from the active registry).
**Verification:** Confirmed that `bazelisk test //server/...` and `bazelisk test //ui:lint` pass cleanly. All tests are hermetic and mimic standard Go best testing practices. No regressions to existing behavior or tests.
