## 2025-05-23 - fasttemplate Optimization
**Learning:** `fasttemplate.ExecuteString` creates an intermediate map and converts all values to strings using `fmt.Sprintf` before execution, causing significant overhead (allocations + CPU). `ExecuteFuncStringWithErr` allows direct writing to the underlying buffer and checking for missing keys during execution, avoiding O(N) pre-scan and intermediate allocations.
**Action:** When using `fasttemplate`, prefer `ExecuteFuncString` or `ExecuteFuncStringWithErr` with a closure to handle values directly, especially for hot paths.
