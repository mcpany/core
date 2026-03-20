# Coverage Intervention Report

* **Target:** `server/pkg/middleware/registry.go` (`GetHTTPMiddlewares` and `GetMCPMiddlewares`)
* **Risk Profile:** The registry file manages the routing and execution order of all middleware components (like auth, caching, rate limiting, DLP) sitting in front of AI request execution. The middleware sorting logic was previously unverified in tests, introducing a significant risk of regressions where critical security or validation middlewares could run out of order (or be bypassed).
* **New Coverage:**
    * Implemented comprehensive Table-Driven tests (`sorts_middlewares_by_priority`) for both HTTP and MCP middleware retrieval functions.
    * The new tests explicitly cover the `sort.Slice` comparison logic, providing multiple `configv1.Middleware` configurations out of order, and ensuring the sorted output respects the specified integer priority boundaries.
    * Added asserting on the sorted arrays implicitly through verifying the expected ordered middleware chain's executed properties. To clear test state safely, implemented `ClearRegistryForTesting()` to reset the package scope `globalRegistry` map instances across concurrent tests.
* **Verification:** `bazelisk test //server/pkg/middleware:middleware_test` verifies all unit tests in the modified package pass properly. Running `bazelisk test //server/...` confirms there are no regressions across the entire suite, including the full integration tests.