# Design Doc: Autonomous Intent Reconciliation (AIR) Hub
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
As agent swarms transition to "Horizontal Teammate" models, multiple sovereign agents often produce conflicting instructions or enter "Negotiation Deadlocks" where they overwrite each other's state on the Blackboard. The AIR Hub provides the authoritative arbitration layer within MCP Any to resolve these conflicts via hardware-attested quorums, ensuring mission-root stability.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-signature "Intent Quorum" for state commits.
    * Detect and resolve circular reasoning dependencies.
    * Provide a standardized "Resolution Result" to all teammates to maintain context parity.
* **Non-Goals:**
    * Replacing the SQLite backend (AIR is a logic layer above the Blackboard).
    * Managing L7 routing (handled by the A2A Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Horizontal Agent Team Orchestrator
* **Goal:** Resolve a state conflict where the "Dev Agent" and "Auditor Agent" simultaneously attempt to modify the system schema with divergent logic.
* **The Happy Path (Tasks):**
    1. Both agents submit conflicting "Intent Fragments" to the AIR Hub.
    2. AIR Hub identifies the conflict and triggers an "Intent Quorum" cycle.
    3. Independent "Monitor" agents provide hardware-attested votes based on mission-root alignment.
    4. AIR Hub selects the winning fragment, signs the "Resolution Result," and commits to the Blackboard.
    5. Teammates receive the update and re-align their reasoning branches.

## 4. Design & Architecture
* **System Flow:**
    * [Agent] -> [AIR Hub Middleware] -> [Quorum Engine] -> [Blackboard (SQLite)]
* **APIs / Interfaces:**
    * `air.submitFragment(intentToken, fragmentHash, data)`
    * `air.getResolution(missionId)` -> Returns winning state.
* **Data Storage/State:**
    * AIR Hub maintains an in-memory "Conflict Graph" and temporary quorum results.

## 5. Alternatives Considered
* **Last-Write-Wins (LWW)**: Rejected as it leads to "Reasoning Regressions."
* **Optimistic Locking**: Rejected due to the high token cost of LLM retry loops.

## 6. Cross-Cutting Concerns
* **Security**: Quorum votes require hardware attestation.
* **Observability**: Resolution logs are surfaced in the "Swarm Topology Widget."

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
