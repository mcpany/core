1. **Target Identification:**
   - **Target:** `server/pkg/tokenizer/tokenizer.go`
   - **Risk Profile:** The `tokenizer` package is highly complex (contains functions like `countTokensInValueSimpleFast` with high cyclomatic complexity) and is a core utility likely used in many places (such as LLM integrations or metric recording) to determine payload sizes and limits. Bugs here could lead to incorrect billing or dropped requests due to miscalculated payload sizes. It currently has multiple branches dealing with various types (`map[string]int`, `map[string]float64`, `[]byte`, cyclic data structures) that have ~0% coverage.
   - **Why this code:** The function `countTokensInValueSimpleFast` has a cyclomatic complexity of 35, and `tokenizer.go` has multiple uncovered functions dealing with deep recursive tokenization. Testing these will improve core robustness.

2. **Test Implementation:**
   - Create comprehensive table-driven tests in `server/pkg/tokenizer/tokenizer_test.go` mimicking the existing test style.
   - Specifically target the uncovered blocks in `countTokensInValueSimpleFast` and `countWordsInValueFast`:
     - `map[string]int`
     - `map[string]int64`
     - `map[string]float64`
     - `map[string]bool`
     - `[]byte`
   - Target the uncovered paths in `countTokensReflect` functions (`countTokensReflectStruct`, `countTokensReflectSlice`, `countTokensReflectMap`), especially cycle detection.
   - Target `countSliceInterfaceSimple` and `countSliceInterfaceRaw` for `[]interface{}`.
   - Target integer length branches in `simpleTokenizeInt64`.
   - The tests will check the exact behavior (the token counts) as expected by the existing logic.

3. **Regression Gate:**
   - Ensure the new tests pass and do not break existing functionality.
   - Run `go test -coverprofile=tmp_cover.out ./server/pkg/tokenizer` to verify the coverage of `countTokensInValueSimpleFast` goes up significantly.
   - Run `make test` and `make lint` across the repo.

4. **Pre Commit Steps:**
   - Ensure proper testing, verification, review, and reflection are done by calling the pre commit instructions tool.

5. **Impact Report Generation:**
   - Write the impact report to `COVERAGE_INTERVENTION_REPORT.md`.
