# Coverage Intervention Report

* **Target:** `server/pkg/storage/sqlite/store.go` and `server/pkg/storage/postgres/store.go`
* **Risk Profile:** The functions `SaveLog` and `GetRecentLogs` lacked test coverage. This logic is considered high risk because it interacts heavily with raw SQL queries, parses complex datatypes (`sql.NullString`), and decodes JSON objects. Failure in `GetRecentLogs` could lead to silent data omission in the UI broadcater rendering logs unavailable or out of order.
* **New Coverage:** Added table-driven/sub-tests to test logging capabilities directly against the database models using existing mocking patterns (`sqlmock` for PostgreSQL and the test database pattern for SQLite). `SaveLog` and `GetRecentLogs` functions are fully executed. The edge case of ensuring correct chronological ascending sorting from a reverse-ordered selection subset is specifically tested. Unmarshaling logic is verified.
* **Verification:** Confirm that `make test` and `make lint` passed cleanly. Also ran `bazelisk test //server/...` which executed successfully.
