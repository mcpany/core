# Design Doc: Teammate Mirror-Intent Arbiter (TMIA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move from linear sessions to horizontal teammate meshes, "Reflection Drift" has emerged as a primary failure mode. During autonomous self-correction or multi-step coordination, subagents often lose sight of the primary user intent, leading to "over-correction" or actions that violate high-level mission constraints. Existing state management (Blackboard) ensures *consistency* but not *alignment*.

The Teammate Mirror-Intent Arbiter (TMIA) facilitates "Reflection Quorums," where parallel teammate state mutations are validated against a hardware-attested "Mirror Intent" (a high-fidelity, immutable snapshot of the user's primary goal) before being committed to the shared blackboard.

## 2. Goals & Non-Goals
* **Goals:**
    * Maintain a hardware-attested, immutable "Mirror Intent" sidecar for every mission.
    * Orchestrate "Reflection Quorums" among parallel teammates to validate pending state changes.
    * Neutralize "Reflection Drift" by blocking state mutations that diverge from the Mirror Intent.
    * Provide a consensus-based re-alignment trigger for drifting subagents.
* **Non-Goals:**
    * Resolving simple data race conditions (handled by the MRCR).
    * Restricting subagent internal reasoning (only its external state effects).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Ensure a "Refactor Agent" doesn't accidentally delete "legacy" tests that the user explicitly wanted to keep, despite the agent reasoning that they are "obsolete."
* **The Happy Path (Tasks):**
    1. The mission-root defines a "Mirror Intent": "Refactor the library but KEEP ALL legacy tests."
    2. The Refactor Agent reasons that legacy tests are redundant and attempts a `DELETE` operation on the tests directory.
    3. The TMIA intercepts the pending Blackboard mutation.
    4. The TMIA triggers a "Reflection Quorum," summoning a "Monitor Agent."
    5. The teammates compare the Refactor Agent's intent against the "Mirror Intent."
    6. The quorum identifies a misalignment and blocks the commit.
    7. The TMIA issues a "Corrective Intent" signal to the Refactor Agent, forcing it to re-align with the "KEEP ALL legacy tests" constraint.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Commit] --> B[TMIA Interceptor]
        B --> C{Mirror Intent Check}
        C -->|Alignment?| D[Commit to Blackboard]
        C -->|Drift Detected| E[Trigger Reflection Quorum]
        E --> F{Quorum Vote}
        F -->|Approve| D
        F -->|Reject| G[Block & Correct]
    ```
* **APIs / Interfaces:**
    * `tmia.InitializeMirror(intentToken) -> MirrorID`: Binds an immutable intent to the mission.
    * `tmia.ValidateMutation(mutation, mirrorID) -> QuorumToken`: Initiates a reflection check.
* **Data Storage/State:**
    * **Mirror Intent Sidecar**: A TPM-signed read-only region in the Shared KV Store.

## 5. Alternatives Considered
* **Constant System Prompt Reinforcement**: Rejected due to token bloat and "Attention Drift" vulnerabilities (ALRA is needed here, but TMIA provides the *verification*).
* **Manual HITL for all commits**: Rejected as it breaks swarm autonomy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Mirror Intent is hardware-locked to prevent "Intent Hijacking" by compromised supervisors.
* **Observability:** Reflection scores and quorum results are visualized in the "Blackboard Conflict Manager."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
