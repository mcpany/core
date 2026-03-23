# Coverage Intervention Report

## Target
* `server/pkg/upstream/mcp/session_registry.go`
* `server/pkg/upstream/mcp/bundle_gc.go`
* `server/pkg/upstream/mcp/stdio_transport.go`

## Risk Profile
These components are critical elements of the `upstream/mcp` integration. `session_registry.go` manages session routing via concurrency primitives (`sync.RWMutex`), which are notoriously difficult to test and prone to data races. `bundle_gc.go` manages the automated background cleanup of temporary resources (bundle directories) using `atomic.Int64` and scheduled goroutines. `stdio_transport.go` handles the raw Inter-Process Communication (IPC) via `stdin`/`stdout`/`stderr` streams, orchestrating the start/stop and monitoring of subprocesses representing tools. Failures in these components could cause deadlocks, resource leaks, data races, or unhandled subprocess crashes respectively. All these files had exactly `0.0%` test coverage before this intervention, indicating a severe blind spot in the codebase's reliability.

## New Coverage
* **`session_registry.go`**: Implemented `TestSessionRegistry_Concurrency` and `TestSessionRegistry_MultipleSessions` which explicitly guard against map concurrent access violations and ensures mapping data consistency between Upstream and Downstream interfaces.
* **`bundle_gc.go`**: Implemented `TestBundleGC_TriggerGCTime` which ensures that the `triggerGC` function accurately respects the interval timer logic (using `CompareAndSwap`), successfully spinning up the background cleanup task when appropriate, and preventing duplicate GC runs.
* **`stdio_transport.go`**: Implemented multiple tests (`TestStdioTransport_ConnectAndReadWrite`, `TestStdioTransport_ConnectError`, `TestStdioTransport_Close`, `TestStdioTransport_ReadBadJSON`) targeting the `Connect`, `Read`, `Write` and `Close` workflows without removing pre-existing coverage like `TestStdioTransport_CaptureStderr`. This guards against edge cases like command execution failures, broken pipe errors, bad JSON payload structures over streams, and improper subprocess shutdown mechanics.

## Verification
* Confirmed that `bazelisk test //server/...` passes securely for all 92 packages.
* Confirmed that all tools (including `pre_commit_instructions`) are completed and pass cleanly.
