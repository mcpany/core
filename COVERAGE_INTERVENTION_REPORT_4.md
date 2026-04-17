# Coverage Intervention Report

* **Target:** `server/pkg/tool/types.go`
* **Risk Profile:** The functions `checkFindInjection`, `checkSQLiteInjection`, `checkMySQLInjection`, `checkPSQLInjection`, `checkSQLKeywords`, and `checkTarInjection` act as core security validations to prevent command and code injections across multiple backends (SQLite, MySQL, PostgreSQL, tar, and find). They had minimal or no direct test coverage, making this a critical, high-risk area. A bug here could lead to Remote Code Execution (RCE) or Data Exfiltration via SQL injection.
* **New Coverage:**
    * `checkFindInjection`: Guards against malicious flags like `-exec`, `-ok`, `-delete`, etc.
    * `checkSQLiteInjection`: Guards against malicious meta-commands like `.shell`, `.system`, `.open`.
    * `checkMySQLInjection`: Guards against `system`, `source`, `INFILE`, and `OUTFILE`.
    * `checkPSQLInjection`: Guards against `\!`, `\o`, and `COPY TO PROGRAM`.
    * `checkSQLKeywords`: Validates input against common SQL injection tactics and keywords (`UNION`, `SELECT`, `--`).
    * `checkTarInjection`: Guards against `--checkpoint-action=exec` flags in `tar`.
* **Verification:** Successfully ran all unit tests for `server/pkg/tool` (`bazelisk test //server/pkg/tool:tool_test`) with no regressions.
