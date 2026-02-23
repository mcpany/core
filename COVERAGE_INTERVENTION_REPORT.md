# Coverage Intervention Report

**Target:** `server/pkg/app/dashboard.go` and `server/pkg/app/api_system.go`

**Risk Profile:**
*   **High Visibility:** The dashboard endpoints directly power the user interface. Failures here lead to a degraded user experience (e.g., broken charts, missing metrics).
*   **Integration Complexity:** `dashboard.go` aggregates data from multiple sources (`TopologyManager`, `ServiceRegistry`, `ToolManager`, `ResourceManager`, `PromptManager`), making it a critical integration point prone to errors if any dependency behaves unexpectedly.
*   **Zero Coverage:** Prior to this intervention, these endpoints had 0% test coverage ("Dark Matter").

**New Coverage:**
*   **`handleDashboardMetrics`:**
    *   Verified Happy Path: Correctly aggregates stats from `TopologyManager` (requests, errors, latency).
    *   Verified Throughput Calculation: Ensures time-windowed throughput (RPS) logic is correct.
    *   Verified Counts: Validates counts from `ServiceRegistry`, `ToolManager`, etc. are correctly propagated to the response.
    *   Verified Method Not Allowed: Ensures non-GET requests are rejected.
*   **`handleSystemStatus`:**
    *   Verified Happy Path: Checks that uptime, version, and warnings are returned correctly.
    *   Verified Warning Logic: Ensures missing API Key generates a security warning.

**Verification:**
*   `go test -v ./server/pkg/app -run TestDashboardMetrics` passed.
*   `go test -v ./server/pkg/app -run TestSystemStatus` passed.
*   `make lint` passed cleanly.
