# Coverage Intervention Report

* **Target:** `server/pkg/upstream/http/http.go`
* **Risk Profile:** The HTTP upstream adapter connects directly to arbitrary external REST/JSON APIs and acts as one of the primary entrypoints for MCP tool generation. It has a cyclomatic complexity of 88 in its `createAndRegisterHTTPTools` method and processes various unmarshaled input types, URL logic, parsing, parameter conversion, and tool evaluations based on configurations. Without adequate coverage of edge cases, configuration misinterpretations could lead to bad downstream calls, panics, incorrect parameter binding, and skipping the evaluation of crucial export policies.
* **New Coverage:**
  * Guarded the `Shutdown` method against nil interface conversions and dereferences, a key stability point when dynamically destroying or hot-reloading configurations.
  * Verified edge case behaviors when parsing `Address` for base URLs gracefully return 0 parsed tools without breaking initialization.
  * Assured the `ToolExportPolicy` evaluation correctly respects "Allowlist" and completely skips internal tools, preventing private tools from mistakenly leaking into discovery.
  * Verified empty or malformed `EndpointPath` structures (e.g. `//`) properly coalesce.
  * Guarded complex internal `structpb` iterations via invalid input schemas gracefully proceeding, rather than panicking the server dynamically.
* **Verification:** `bazelisk test //server/pkg/upstream/http/...` and `bazelisk test //server/...` both completed with 100% green tests showing hermetic test execution with no broken assumptions in the overall architecture.

## Update

* Revised the tests into a proper Go table-driven suite matching the repository's convention (`TestCoverageIntervention_RegisterEdgeCases` & `TestCoverageIntervention_Shutdown`).
* Ensured the new test file `coverage_intervention_test.go` compiles and successfully passes via hermetic execution (`bazelisk test //server/pkg/upstream/http/...`).
* Checked that the new `COVERAGE_INTERVENTION_REPORT.md` file correctly encompasses the information and is added appropriately.
