# Coverage Intervention Report

* **Target:** `server/pkg/app/api_users.go` (`handleUsers` and `handleUserDetail`)
* **Risk Profile:** These endpoints manage user identity, authentication hashing, profile updates, and role assignments. The code processes password hashes, enforces RBAC (Role-Based Access Control) to prevent privilege escalation (IDOR), and creates users securely. The endpoints exhibited low test coverage (`~50.2%`), which poses a significant risk to the integrity of system access management and identity provisioning.
* **New Coverage:**
    * Implemented comprehensive table-driven tests (`TestHandleUsers_TableDriven` and `TestHandleUserDetail_TableDriven`) in `server/pkg/app/api_users_extra_test.go`.
    * Coverage now rigorously tests edge cases for HTTP methods, authentication levels, and data formatting.
    * Guarded logic paths include:
        * Rejection of `MethodNotAllowed` verbs and lack of admin roles for `GET` and `POST` operations on `/users`.
        * Missing ID fields and existing user conflict handling on creation.
        * Missing user authentication context detection or forbidden profiles in detail retrieval/updates.
        * Incorrect/missing IDs in the `PUT` request parameters causing `400 Bad Request`.
* **Verification:** `bazel test //server/pkg/app:app_test` and `bazel test //server/...` confirmed that tests pass correctly. There are no downstream functional regressions, and all tests remain hermetic and conform to established patterns (e.g. Google Standard table-driven testing).

