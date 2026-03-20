# Coverage Intervention Report

* **Target:** `server/pkg/storage/sqlite/db.go`
* **Risk Profile:** This file was selected because it is a critical initialization component for the entire SQLite storage backend. With a cyclomatic complexity of 12 and zero test coverage, any failure to correctly initialize the schema, set proper directory permissions, or apply essential performance and safety PRAGMAs (like WAL mode, synchronous=NORMAL, and a strict connection limit) could lead to catastrophic data corruption or application startup failure. Testing this ensures the fundamental database layer is rock solid.
* **New Coverage:**
  - **Happy Path (`TestNewDB_Success`):** Guards the initialization logic, confirming the database file is correctly created when given a valid path. It also asserts that the correct PRAGMAs are executed successfully, verifying `journal_mode` is `wal`, `synchronous` is 1 (`NORMAL`), `busy_timeout` is `5000`, and `db.SetMaxOpenConns(1)` logic.
  - **Edge Case (`TestNewDB_Failure_Mkdir`):** Guards the failure path when directory creation fails (e.g., when the target path is blocked by an existing file).
  - **Schema Verification (`TestInitSchema_TablesExist`):** Ensures that all 10 required tables (`upstream_services`, `global_settings`, `secrets`, `users`, `profile_definitions`, `service_collections`, `user_tokens`, `credentials`, `service_templates`, `logs`) and the crucial index (`idx_logs_timestamp`) are properly created by the `initSchema` script.
* **Verification:** Confirmed that tests pass cleanly for the file, and the changes are completely isolated and hermetic without affecting existing application code or causing regressions in the rest of the codebase.
