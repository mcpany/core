# Design Doc: Phase-Bound State Sealing (PBSS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent missions move from linear sessions to complex, multi-phase workflows (e.g., Research -> Code -> Test -> Deploy), the risk of "Context Leaking" or "Phase-Splicing" increases. Specialists in later phases should not be able to probe sensitive state fragments from earlier phases that are no longer relevant to their task.

Phase-Bound State Sealing (PBSS) ensures mission-phase integrity by cryptographically sealing state snapshots upon phase completion. This prevents subagents from accessing legacy data and mitigates the risk of a compromised specialist in a late-stage phase exfiltrating sensitive data from an earlier phase.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically seal Blackboard state fragments upon transition to a new mission phase.
    * Provide cryptographic proof that legacy phase state is inaccessible to active subagents.
    * Support hardware-locked phase transitions (integrated with MBHL).
    * Enable "Intent-Gated" resumption of specific sealed fragments for supervisors.
* **Non-Goals:**
    * Replacing the Shared KV Store (PBSS is a lifecycle manager for it).
    * Permanent deletion of data (Sealing is for active subagent access control).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Swarm Orchestrator
* **Primary Goal:** Ensure that a "Deployment Specialist" subagent cannot access the "API Key Discovery" phase's reasoning traces.
* **The Happy Path (Tasks):**
    1. The Mission Lead agent signals the completion of Phase 1 (Credential Discovery).
    2. PBSS intercepts the transition and generates a hardware-attested snapshot of Phase 1 state.
    3. PBSS seals the Phase 1 fragments, making them read-only and invisible to non-supervisor agents.
    4. The Mission transitions to Phase 2 (Code Generation).
    5. A Code Generation subagent attempts to read a Phase 1 state fragment.
    6. MCP Any blocks the request, returning a "Phase Sealed" integrity error.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Phase Transition Signal] --> B[PBSS Controller]
        B --> C[State Snapshot Engine]
        C --> D[Hardware Key Mint (TPM)]
        D --> E[Sealed State Vault]
        E --> F{Active Request?}
        F -- Valid Phase --> G[Grant Access]
        F -- Legacy Phase --> H[Block & Log]
    ```
* **APIs / Interfaces:**
    * `SealPhase(missionID, phaseID) error`
    * `GetSealedFragment(missionID, fragmentID, supervisorToken) (Fragment, error)`
* **Data Storage/State:** Sealed snapshots are stored in the mission's persistent SQLite sidecar, encrypted with phase-specific hardware keys.

## 5. Alternatives Considered
* **Manual State Deletion:** Rejected due to the risk of "Amnesia" where supervisors need to audit legacy steps.
* **Namespace Isolation:** Rejected as insufficient against subagents that can discover namespace patterns. PBSS uses cryptographic sealing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PBSS implements the principle of least privilege across time (temporal isolation).
* **Observability:** Phase seal events and blocked access attempts are logged in the Mission Continuity Provider (MRCP) audit trail.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
