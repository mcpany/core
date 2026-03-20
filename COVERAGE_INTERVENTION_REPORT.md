# Coverage Intervention Report

* **Target:** `server/pkg/app/api_skills.go` (`handleSkills` and `handleSkillDetail`)
* **Risk Profile:** These endpoints handle HTTP operations mapping to internal skill operations in `server/pkg/app`. It exhibited low coverage and high Cyclomatic Complexity. Leaving these endpoints untested posed a risk of regressions, especially related to the security and management of skills (creation, retrieval, updates, asset routing, etc.).
* **New Coverage:**
    * I implemented a comprehensive table-driven test `TestHandleSkills` which tests `GET` (List), `POST` (Create with Success, Invalid Body, and Creation Error scenarios), and default method handler (Method Not Allowed).
    * I implemented a comprehensive table-driven test `TestHandleSkillDetail` which tests `GET` (Retrieve with Success and Not Found scenarios), `PUT` (Update with Success, Invalid Body, and Update Error scenarios), `DELETE` (Delete with Success and Not Found scenarios), default method handler (Method Not Allowed), missing skill names in the request URL, and correct asset routing (delegation to `handleUploadSkillAsset`).
    * The new coverage mimics the Google Standard Table-Driven Test pattern present in `TestHandleUploadSkillAsset`.
* **Verification:** `bazelisk test //server/pkg/app:app_test` confirms tests pass correctly without modifying underlying functionality. Running `bazelisk test //server/...` confirms there are no new regressions.
