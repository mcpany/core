# Coverage Intervention Report

**Top 10 High-Risk Components Identified:**
1. `server/pkg/app/server.go: (*Application).runServerMode` (Cyclomatic Complexity: 120)
2. `server/pkg/upstream/http/http.go: (*Upstream).createAndRegisterHTTPTools` (Cyclomatic Complexity: 88)
3. `server/pkg/tool/types.go: (*CommandTool).Execute` (Cyclomatic Complexity: 66)
4. `server/pkg/storage/sqlite/store_test.go: TestStore` (Cyclomatic Complexity: 64)
5. `server/pkg/tool/types.go: (*LocalCommandTool).Execute` (Cyclomatic Complexity: 63)
6. `server/pkg/app/server.go: (*Application).Run` (Cyclomatic Complexity: 63)
7. `server/pkg/util/secrets.go: resolveSecretImpl` (Cyclomatic Complexity: 49)
8. `server/pkg/tokenizer/tokenizer_fastpath_test.go: TestCountTokensInValue_FastPathConsistency` (Cyclomatic Complexity: 44)
9. `server/pkg/app/server.go: (*Application).reconcileServices` (Cyclomatic Complexity: 43)
10. `server/pkg/tool/types.go: (*OpenAPITool).Execute` (Cyclomatic Complexity: 40)

* **Target:** `server/pkg/tokenizer/tokenizer.go`
* **Risk Profile:** This file contains highly complex and optimized tokenization logic used across the system for cost calculations, context window management, and rate limiting. The fast path functions (`countTokensInValueSimpleFast` and `countTokensInValueWordFast`) have a high cyclomatic complexity (35) due to exhaustive type switches. While they had ~97% coverage, critical edge cases involving small values evaluating to < 1 token and cyclic pointer reflection references were untested, posing a risk of subtle token accounting bugs or infinite recursion errors.
* **New Coverage:**
  - Added table-driven test cases for explicitly small inputs (`len(text) < 4`, small integers `int8` through `int64`, and small floats) that trigger the `count < 1` bounds correction.
  - Added tests covering the empty slice/map handling.
  - Added tests for `map` and `slice` data types exhibiting cyclical references to ensure the `visited[ptr]` short-circuiting logic accurately intercepts cyclic resolution and correctly errors out ("cycle detected").
* **Verification:** Confirmed that `make test` and `make lint` passed cleanly.
