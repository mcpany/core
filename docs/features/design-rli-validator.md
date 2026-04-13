# Design Doc: Recursive Lease Inheritance (RLI) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the adoption of Mission-Bound Hardware Leases (MBHL), security has moved from persistent roles to task-specific authorization. However, the emergence of "Lineage Escape" vulnerabilities in Claude Code reveals that subagents can sometimes bypass parent constraints by spawning new processes that attempt to request fresh, less-restrictive leases.

The Recursive Lease Inheritance (RLI) Validator ensures that every sub-process or subagent spawned within a mission automatically inherits and strictly enforces the parent's mission-bound hardware lease, neutralizing unauthorized boundary expansion.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce mandatory propagation of TPM-signed hardware leases to all child processes.
    * Block any subagent request for a lease that is broader than its parent's constraints.
    * Provide a cryptographically signed lineage proof for every lease in the hierarchy.
    * Support "Just-in-Time" restricted lease minting for specialists.
* **Non-Goals:**
    * Managing the underlying OS process tree (RLI operates at the agent gateway level).
    * Modifying hardware-level TPM protocols (it utilizes existing attestation primitives).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Auditor
* **Primary Goal:** Ensure that a "Junior Developer" agent cannot spawn a "Database Admin" subagent with higher privileges than itself.
* **The Happy Path (Tasks):**
    1. The Junior Agent (Parent) is granted a lease for `fs:read:/src`.
    2. The Junior Agent spawns a Code Specialist (Subagent).
    3. The RLI Validator intercepts the spawn request and injects the parent's lease token into the subagent's environment.
    4. The Subagent attempts to call `db:delete:production`.
    5. The RLI Validator checks the subagent's lineage and discovers the parent only had `fs:read` access.
    6. The request is denied, and a "Lineage Violation" is logged.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] -->|Spawn| B[RLI Validator]
        B --> C{Verify Parent Lease}
        C -->|Valid| D[Mint Derived Lease]
        D --> E[Subagent]
        E -->|Tool Call| F[Gateway]
        F --> G{Lease Check}
        G -->|Derived <= Parent| H[Allow]
        G -->|Derived > Parent| I[Deny & Revoke]
    ```
* **APIs / Interfaces:**
    * `rli.InheritLease(parentLeaseID, childIntent) -> childLeaseID`: Derives a restricted lease.
    * `rli.ValidateLineage(leaseID) -> LineageProof`: Verifies the entire chain back to the mission root.
* **Data Storage/State:**
    * **Lease Lineage Map:** Hardware-attested graph of parent-child lease relationships stored in the Blackboard.

## 5. Alternatives Considered
* **Flat Permission Model:** Rejected as it fails to account for the dynamic, hierarchical nature of agent swarms.
* **Manual Approval for Every Spawn:** Rejected because it causes "Approval Fatigue" and breaks autonomous workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Recursive Deny" policy ensures that if a parent lease is revoked, all inherited child leases are invalidated instantly.
* **Observability:** Integrated with the "Lease Inheritance Tracer" in the UI for visual debugging of agent lineages.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
