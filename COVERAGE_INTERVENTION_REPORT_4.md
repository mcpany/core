# Coverage Intervention Report

* **Target:** `server/pkg/lifecycle/reaper.go` (`RegisterSubagent`)
* **Risk Profile:** The `RegisterSubagent` function controls the registration of subagents onto active leases, serving as the bridge tying session IDs to their lifecycle contexts. It is critical path logic for preventing zombie processes and limiting resource utilization, but had zero test coverage prior to this intervention. Its Cyclomatic Complexity is relatively low, but its risk context (managing critical concurrency state in a lock-protected struct) puts it firmly in High Risk.
* **New Coverage:** Added testing for the `RegisterSubagent` method mimicking the existing table-driven/direct test styles. The new tests verify:
    * **Happy Path:** Registration successfully associates a session with an `ACTIVE` lease without errors.
    * **Edge Case (Non-existent Lease):** Prevents a subagent from registering to a lease intent that does not exist in the registry.
    * **Edge Case (Invalid State):** Prevents a subagent from joining a lease intent that is not `ACTIVE` (e.g., manually transitioned to `PRUNED` or automatically transitioned to `EXPIRED`).
* **Verification:** Confirmed that `go test -v ./pkg/lifecycle/...` ran completely green across the `lifecycle` package. Ran `make lint` clean at repository root.
