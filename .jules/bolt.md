## 2025-05-23 - fasttemplate Optimization
**Learning:** `fasttemplate.ExecuteString` creates an intermediate map and converts all values to strings using `fmt.Sprintf` before execution, causing significant overhead (allocations + CPU). `ExecuteFuncStringWithErr` allows direct writing to the underlying buffer and checking for missing keys during execution, avoiding O(N) pre-scan and intermediate allocations.
**Action:** When using `fasttemplate`, prefer `ExecuteFuncString` or `ExecuteFuncStringWithErr` with a closure to handle values directly, especially for hot paths.

## 2025-12-17 - Regex Compilation Bottleneck in Policy Hooks
**Learning:** `regexp.MatchString` compiles the regex on every call. In the policy hook execution path, which runs for every tool call, this caused repeated compilation of policy regexes.
**Action:** Pre-compile regexes and store them in the struct. For dynamic configurations, cache the hooks/compiled regexes in the configuration object or manager to avoid re-creation on every request.
