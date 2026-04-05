# Design Doc: Recursive Lease Delegation (RLD) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the transition to complex, horizontal "Agent Teams" (e.g., Claude Code v3.3.0), subagents often need to spawn specialized sub-processes to execute discrete tasks. Currently, these sub-processes either inherit the full capabilities of the parent (violating Least Privilege) or fail due to lack of authority (Scope Exhaustion).

The Recursive Lease Delegation (RLD) Manager provides a standardized mechanism for agents to delegate granular, subsetted hardware leases to their descendants. This ensures that every process in a deep delegation chain has exactly the authority it needs, backed by the mission-root's TPM signature, without redundant manual attestation.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hierarchical subsetting of hardware-attested capability leases.
    * Maintain cryptographic lineage for all delegated leases back to the mission-root.
    * Prevent "Scope Exhaustion" by allowing dynamic refinement of leases.
    * Automate lease revocation when a sub-mission or process terminates.
* **Non-Goals:**
    * Providing cross-platform lease translation (handled by CFAT).
    * Bypassing initial user-in-the-loop attestation for the mission-root.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Securely delegate a "Read-Only" subset of a parent's "Filesystem:Full" lease to a 4th-level sub-process.
* **The Happy Path (Tasks):**
    1. Parent agent (with `fs:write:/project`) spawns a specialist search subagent.
    2. Parent agent calls `rld.DelegateLease()` with a scope restricted to `fs:read:/project/src`.
    3. RLD Manager generates a new TPM-signed lease fragment linked to the parent's lineage.
    4. Subagent receives the fragment and uses it to authorize its tool calls.
    5. When the search task completes, the RLD Manager automatically revokes the sub-lease, returning the scope to the parent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root TPM] --> L1[Parent Lease]
        L1 --> RLD[RLD Manager]
        RLD --> L2[Sub-Lease A: Read-Only]
        RLD --> L3[Sub-Lease B: Network-Only]
        L2 --> SA[Specialist Agent]
    ```
* **APIs / Interfaces:**
    * `rld.DelegateLease(parentLeaseID, childScope) -> SubLeaseID`: Creates a restricted descendant lease.
    * `rld.VerifyLease(leaseID) -> LineageProof`: Validates a lease against the TPM root.
    * `rld.RevokeLease(leaseID)`: Immediate termination of a delegated scope.
* **Data Storage/State:**
    * **Lease Lineage Map:** Tree-based structure tracking the parent-child relationships of active leases.
    * **TPM Counter Registry:** Synchronized with the AMRA Hub to prevent replay of expired leases.

## 5. Alternatives Considered
* **Flat Intent Inheritance:** Rejected as it violates the principle of Least Privilege and increases the impact of a compromised specialist agent.
* **Per-Process User Re-attestation:** Rejected due to "Approval Fatigue" and the breakdown of autonomous swarm velocity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RLD fragments are mathematically incapable of exceeding the parent's scope. Lineage-aware validation is enforced at every tool call by the DPG.
* **Observability:** Integrated with the "Hierarchical Trust Monitor" for real-time visualization of the lease tree and delegation depth.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
