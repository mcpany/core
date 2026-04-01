1. *Write tests for `server/pkg/app/server.go` covering config diff functions.*
   - Target functions: `generateConfigDiff` and `captureConfigSnapshot`.
   - Add a test file `server_diff_test.go` or add to `server_test.go` or `server_init_test.go`.
2. *Run the specific tests to verify coverage.*
   - Use bazel test and coverage output to confirm it increases coverage on lines 1283-1360.
3. *Run all tests.*
   - Ensure the new tests do not break any existing test, according to the zero-harm principle.
4. *Write tests for `server/pkg/config/store.go` and `server/pkg/config/validator.go` if needed.*
   - The config ones might be more focused. In validator.go, test `validateMtlsAuth` file handling logic.
   - For store.go, test the `replaceEnvVariables` paths handling missing and restricted env variables.
   - Run tests to check coverage.
5. *Pre-commit checks*
   - Will verify tests, and run test loops using pre commit checks tool.
6. *Create the report*
   - Update `COVERAGE_INTERVENTION_REPORT.md` (or append to a new one if requested).
