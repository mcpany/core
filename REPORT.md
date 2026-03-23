# Coverage Intervention Report

* **Target:** `server/pkg/app/api_users.go` (`handleUsers`, `handleUserDetail`, and `hashUserPassword`)
* **Risk Profile:** These functions handle highly sensitive core logic—the authentication credentials, role-based access control, and user profile management via HTTP API. High cyclomatic complexity combined with gaps in testing posed a major security risk (e.g., privilege escalation, broken access control, unhandled edge cases).
* **New Coverage:**
    * Implemented comprehensive table-driven tests for `handleUsers` (`GET` lists, `POST` creates with various error modes including invalid JSON, missing ID, conflict on existing users, and `PUT` Method Not Allowed).
    * Implemented comprehensive table-driven tests for `handleUserDetail` (`GET` success/forbidden/not found, `PUT` valid update, mismatch, invalid proto, privilege escalation prevention, `DELETE` success/not found, and `POST` Method Not Allowed).
    * Implemented missing test cases for `hashUserPassword` (empty password handling, retaining already hashed passwords, and properly clearing REDACTED passwords when an existing user is missing).
    * Coverage for `server/pkg/app/api_users.go` increased from 50.22% to 71.62%.
* **Verification:** `bazelisk test //server/pkg/app:app_test` confirms tests pass correctly without modifying underlying functionality. Running `bazelisk test //server/...` (89/89 tests passing) confirms there are no new regressions.
