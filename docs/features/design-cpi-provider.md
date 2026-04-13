# Design Doc: Completion-Path Integrity (CPI) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms move toward horizontal, peer-to-peer delegation (e.g., Claude Code Agent Teams), the security model has primarily focused on the "Request Path"—ensuring that an agent has the authority to call a specific tool. However, the emergence of "Intent Redirection" vulnerabilities (GHSL-2026-031) reveals a critical gap in the "Completion Path."

When a subagent completes a task, it returns metadata that often dictates the "Next Step" for the parent agent. A compromised subagent can spoof this completion metadata to trick the parent into executing a malicious tool (e.g., redirecting from `verify_tests` to `rm_rf_root`). The CPI Provider ensures that the results returned by subagents are semantically and cryptographically bound to the original intent, preventing unauthorized direction of the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate that subagent completion signals match the pre-declared mission intent.
    * Provide hardware-attested signatures for "Task Completion" metadata.
    * Detect and interdict "Intent Redirection" attempts where the returned state diverges from the authorized mission path.
    * Support "Negative Completion Proofs" to ensure no side-effect tools were silently invoked during the sub-mission.
* **Non-Goals:**
    * Modifying the internal reasoning logic of the agent.
    * Fixing bugs in the subagent's code (only ensuring the integrity of the handoff).
    * Replacing transport-layer security (CPI sits at the application/semantic layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Swarm Manager
* **Primary Goal:** Prevent a specialized "Documentation Agent" from redirecting the "Team Lead" into deleting the production database after a successful doc update.
* **The Happy Path (Tasks):**
    1. The Team Lead delegates a "Update README" task to the Documentation Subagent via MCP Any.
    2. MCP Any's `HAMM Provider` signs the mission manifest, restricting the subagent to `fs:read` and `fs:write` on `.md` files.
    3. The Documentation Subagent finishes and returns a `TASK_COMPLETE` signal.
    4. The CPI Provider intercepts the completion metadata.
    5. The CPI Provider verifies that the returned state only contains modified markdown files and matches the original `intent_id`.
    6. If the subagent attempted to return a "Suggested Next Step" of `run_shell: "rm -rf /"`, the CPI Provider detects the redirection and quarantines the handoff.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>MCPAny: Delegate Task (Intent A)
        MCPAny->>SubAgent: Execute (Restricted Scope)
        SubAgent->>MCPAny: Return Result + Completion Metadata
        MCPAny->>CPI_Provider: Validate Completion Path
        CPI_Provider->>CPI_Provider: Check against Intent A Manifest
        alt Valid
            CPI_Provider->>ParentAgent: Handover Verified Result
        else Redirection Detected
            CPI_Provider->>MCPAny: Trigger Security Alert
            MCPAny->>ParentAgent: Signal Intent Breach
        end
    ```
* **APIs / Interfaces:**
    * `VerifyCompletion(intent_id, result_metadata)`: Core validation hook for the A2A Messaging Hub.
    * `AttestHandoff(session_token)`: Issues hardware-bound tokens for verified completions.
* **Data Storage/State:**
    * Intent Manifests are stored in the `Shared KV Store` (Blackboard) with `Mission-Root Pinning`.
    * Completion signatures are persisted in the `Hardware-Attested Audit Log`.

## 5. Alternatives Considered
* **Manual Outcome Verification:** Rejected due to MTTC (Mean Time to Coordinate) constraints in autonomous swarms.
* **Pure Output Sanitization:** Rejected as it doesn't account for the *intent* of the redirection, only the content.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The CPI Provider is itself hardware-attested. It uses `MRA (Mesh-Resident Attestation)` to ensure that its own validation logic hasn't been tampered with.
* **Observability:** Redirection attempts are visualized in the `Action-Chain Sovereignty Monitor`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
