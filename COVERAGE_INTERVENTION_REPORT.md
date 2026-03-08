# Coverage Intervention Report

* **Target:** `server/pkg/upstream/mcp/bundle_local_transport.go` (and `server/pkg/upstream/mcp/bundle_local_transport_test.go`)
* **Risk Profile:** This code is part of the core tool registration and command execution flow, acting as a gateway that spawns local processes via `exec.CommandContext`. Because it directly executes processes on the host based on configuration, the consequences of bugs here (such as unhandled errors, process leaks, or failures to capture stdio appropriately) represent a high risk. It had 0% test coverage before this intervention.
* **New Coverage:** The following logic paths are now guarded with proper test assertions:
  * **Success Path (`Connect`):** Verifies that `Connect()` correctly instantiates a `StdioTransport`, spawns the desired process, and successfully establishes a two-way JSON-RPC connection over Stdio.
  * **Error Handling (Process startup failure):** Verifies that if a non-existent executable is specified, `Connect()` gracefully fails with an "executable file not found" error rather than crashing or hanging.
  * **Error Handling (Stderr capture on failure):** Verifies that if the target process prints to `stderr` and exits with an error code, the returned error captures and includes the `stderr` output (mimicking the `StdioTransport` behavior).
  * **Edge Case (Early exit):** Tests the case where the target process exits immediately (e.g., executing `ls` on a nonexistent file) to ensure early termination is correctly handled.
* **Verification:** Confirmed that `make test` and `make lint` both passed cleanly without regressions to the broader test suite.