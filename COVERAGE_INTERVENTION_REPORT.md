# Coverage Intervention Report

* **Target:** `server/pkg/app/api_users.go` (`handleUsers` and `handleUserDetail`)
* **Risk Profile:** These endpoints handle critical identity management and role-based access control functionalities for user creation, updating, and listing. Prior coverage was about 50%, missing vital verification for edge cases like invalid JSON representations, missing payload wrappers (e.g. `{"user": ...}` vs raw inputs), missing IDs, and importantly Privilege Escalation (IDOR) attacks by users trying to manipulate their own roles via `PUT` requests. Leaving these endpoints with low coverage poses extreme security and access risks.
* **New Coverage:**
    * Implemented a comprehensive table-driven test `TestHandleUsers_Api` for `GET` (success and authorization denials) and `POST` (success wrappers vs direct mapping, conflict handling, JSON validation, password hashing & redaction validation, default error handler).
    * Implemented a comprehensive table-driven test `TestHandleUserDetail_Api` testing `GET`, `PUT`, `DELETE` operations against correct users. Specifically tested security gates ensuring non-admins cannot `GET`, `PUT`, or `DELETE` target users apart from themselves, and explicitly asserting that IDOR attacks via `PUT` targeting role escalation correctly ignores the changes from non-admins but works for admins.
    * The code style strictly adheres to existing Google-styled table-driven standards found in `api_users_test.go` while fixing missing fields tests on JSON Unmarshaling bounds (`DiscardUnknown: true`).
* **Verification:** `npx @bazel/bazelisk test //server/pkg/app:app_test` confirms all handler edge cases act correctly. Running `npx @bazel/bazelisk test //server/...` ensures zero integration regression has happened.
