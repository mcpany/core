1. **Analyze existing tests & code**:
   - Analyzed codebase and found out that `autogen.go`, `crewai.go`, and `openclaw.go` tests were missing.

2. **Create new test files**:
   - `src/interop/autogen_test.go`
   - `src/interop/crewai_test.go`
   - `src/interop/openclaw_test.go`

3. **Verify files were created and run tests**:
   - `bazelisk test //src/interop/...`
