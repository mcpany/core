# Design Doc: Recursive Intent Deconstruction (RID) Hub
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms become deeper and more autonomous, single-hop attestation (verifying just the immediate parent) is proving insufficient to detect cumulative "Intent Drift." A subagent three levels deep may pass all individual security checks while performing actions that fundamentally diverge from the user's original "Mission Root."

The Recursive Intent Deconstruction (RID) Hub is needed to provide deep semantic auditing of sub-missions by recursively deconstructing the reasoning path into its atomic intent fragments, which are then hardware-attested against the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Recursively deconstruct agent reasoning paths into atomic intent fragments.
    * Enforce hardware-attested semantic consistency across the entire delegation lineage.
    * Provide a deterministic mechanism to detect intent hijacking in deep swarms.
    * Integrate with the RIV Provider for lineage-aware proofs.
* **Non-Goals:**
    * Directly managing tool execution (handled by the tool gateway).
    * Providing real-time attention locking (handled by HAAL).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Level Swarm Orchestrator
* **Primary Goal:** Ensure that a deeply nested subagent remains mission-aligned without manual review of every hop.
* **The Happy Path (Tasks):**
    1. A root agent delegates a task to Specialist A.
    2. Specialist A spawns Sub-Specialist B.
    3. RID Hub intercepts the spawn request and deconstructs Specialist A's reasoning for the delegation.
    4. RID Hub verifies the deconstructed intent fragments against the Mission Root.
    5. RID Hub issues a "Lineage-Attestation Token" to Sub-Specialist B.
    6. When B attempts a tool call, the RID Hub verifies the entire token chain back to the Mission Root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Spawn Request] --> B[RID Hub]
        B --> C[Recursive Deconstructor]
        C --> D[Lineage Verification Engine]
        D --> E{Aligned with Root?}
        E -- Yes --> F[Issue Lineage-Attestation Token]
        E -- No --> G[Block & Alert Mission Root]
        H[Hardware-Attested Mission Root] --> D
        I[RIV Provider] --> F
    ```
* **APIs / Interfaces:**
    * `rid.DeconstructLineage(sessionID) -> IntentChain`: Deconstructs the reasoning history.
    * `rid.VerifyLineage(intentChain, missionRoot) -> bool`: Validates the chain against the root.
* **Data Storage/State:**
    * **Intent Fragment Store:** TPM-protected storage for attested intent fragments.

## 5. Alternatives Considered
* **Flattened Permissions:** Rejected because it loses the semantic context of *why* an action was authorized.
* **Full Monologue Review:** Rejected due to prohibitive token costs and reasoning latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage tokens must be hardware-bound to prevent "Token Splicing" by compromised agents.
* **Observability:** Integrated with the "Mesh-Resident Lineage Tracker" for visual auditing.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
