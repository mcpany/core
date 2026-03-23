# Coverage Intervention Report

* **Target:** `server/pkg/app/api.go` (`handleCollectionApply`)
* **Risk Profile:** This endpoint handles HTTP operations for applying a collection of upstream services to the server. It exhibited zero coverage and moderate complexity (iterating over services, conditional validation, safety checks, and saving state). Leaving this endpoint untested posed a risk of regressions, especially related to the security of allowing unsafe services to be configured via the API without proper `MCPANY_ALLOW_UNSAFE_CONFIG` environment controls.
* **New Coverage:**
    * I implemented a comprehensive table-driven test `TestHandleCollectionApply` in `server/pkg/app/api_collection_apply_test.go`.
    * The test covers:
        * `GET` (Method Not Allowed)
        * `POST` with non-existent collection (Not Found)
        * `POST` with a Safe Service (Happy Path - Success)
        * `POST` with an Unsafe Service where `MCPANY_ALLOW_UNSAFE_CONFIG=true` (Allowed)
        * `POST` with an Unsafe Service where `MCPANY_ALLOW_UNSAFE_CONFIG=false` (Blocked and Skipped)
        * `POST` with an Invalid Service (Skipped)
    * The new coverage mimics the Google Standard Table-Driven Test pattern present in `server/pkg/app`.
* **Verification:** `bazelisk test //server/pkg/app:app_test` confirms tests pass correctly without modifying underlying functionality. Running `bazelisk test //server/...` confirms there are no new regressions.
