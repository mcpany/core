# Impact Report: Coverage Intervention

## Target
`server/pkg/app/template_manager.go`

## Risk Profile
This component was selected because it manages the **persistence and lifecycle of service templates**, which involves:
*   **Data Integrity:** Reading and writing configuration files (`templates.json`).
*   **Serialization:** Complex Protocol Buffer to JSON marshaling/unmarshaling (`protojson`).
*   **Concurrency:** Shared state protected by `sync.RWMutex`, accessed by multiple API handlers.
*   **Zero Coverage:** It had no dedicated tests prior to this intervention.

## New Coverage
A robust test suite (`server/pkg/app/template_manager_test.go`) has been implemented, covering the following logic paths:
1.  **Initialization:**
    *   Creating a new manager in a fresh directory.
    *   Handling missing directories (auto-creation).
    *   Resilience against corrupt data files (starts empty, logs error).
2.  **CRUD Operations:**
    *   **Create:** Adding new templates.
    *   **Read:** Listing templates.
    *   **Update:** Updating existing templates by ID (primary key) and Name (fallback).
    *   **Delete:** Removing templates by ID or Name.
3.  **Persistence:**
    *   Verifying that changes are written to disk and can be reloaded after a restart.
4.  **Concurrency:**
    *   Stress testing with concurrent readers and writers to ensure thread safety and absence of race conditions.

## Bug Fix
During test implementation, a **critical bug** was discovered in `SaveTemplate`:
*   **Issue:** Updates to existing templates were failing if the `Name` field was changed, resulting in duplicate entries instead of an in-place update.
*   **Fix:** The logic was corrected to prioritize ID matching before falling back to Name matching.

## Verification
*   **New Tests:** `go test -v github.com/mcpany/core/server/pkg/app -run TestTemplateManager` -> **PASS**
*   **Regression:** `go test ./server/pkg/app/...` -> **PASS**
