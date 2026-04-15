# Coverage Intervention Report

* **Target:** `server/pkg/app/api.go`
* **Risk Profile:** This file contains critical security and functional validation logic (`readBodyWithLimit`, `checkFilesystemAccess`, `checkCommandAvailability`) for incoming payloads, filesystems paths and tool executions. Previously, they had no coverage, making the system highly vulnerable to path traversal, unvalidated input execution, DOS by large request payloads, or un-executable tool calls crashing the handler. The complexity is high as they serve as the gateway to execution boundaries.
* **New Coverage:**
  - `TestReadBodyWithLimit`: Verified correct handling of payloads under limits, and correct maxBytesError bubbling and HTTP 413 responses for payloads over limit.
  - `TestCheckFilesystemAccess`: Verified both the happy path (existing) and error paths (non-existent).
  - `TestCheckCommandAvailability`: Exhaustively guards combinations of empty commands, commands in `PATH` vs missing from `PATH`, absolute paths existing vs missing, and working directory validations (including invalid files or paths passed as directories).
* **Verification:** Run `bazelisk test //server/pkg/app:app_test` confirming the target test subsets (`TestReadBodyWithLimit|TestCheckFilesystemAccess|TestCheckCommandAvailability`) all passed.
