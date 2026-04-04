# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-04-04

## 1. Context and Scope
As agents move from linear sessions to horizontal teammate meshes, the risk of "Capability Squatting" has become a critical failure point. Standard session-based permissions are too coarse, allowing subagents to retain high-trust capabilities (e.g., shell access, production DB writes) long after their specific sub-task is complete.

The HLML Provider introduces a hardware-locked, task-specific lease mechanism. Every capability is issued as a TPM-signed lease that is cryptographically bound to a unique mission-root task ID. This ensures that privileges are automatically invalidated by the hardware root of trust upon task completion, neutralizing persistent privilege escalation in deep or horizontal swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) capability leases.
    * Bind every lease to a specific, unique Task ID and Mission Root.
    * Implement automated, hardware-enforced revocation upon task termination.
    * Support "Cascading Invalidation" for child tasks.
* **Non-Goals:**
    * Managing the business logic of task decomposition (handled by the Orchestrator).
    * Providing long-term persistent storage for expired leases.
    * Replacing existing A2A handshakes; it hardens the outcome of those handshakes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Ensure a specialist "Refactor Agent" only has write access to the specific repository subdirectory for the duration of its refactoring sub-task.
* **The Happy Path (Tasks):**
    1. Parent Agent proposes a "Refactor CSS" sub-task to the specialist.
    2. HLML Provider generates a TPM-signed lease for "fs:write:/src/styles" bound to Task-404.
    3. Specialist agent performs the work using the lease as its hardware-attested authorization.
    4. Specialist agent signals task completion to the MRCP (Mission-Root Continuity Provider).
    5. HLML Provider receives the completion signal and signals the hardware TPM to invalidate the lease signature.
    6. Any subsequent attempt to use the lease is rejected at the kernel/enclave level.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Orchestrator] -->|Task Start| B[HLML Provider]
        B -->|Request Signature| C[TPM / Secure Enclave]
        C -->|Signed Lease| B
        B -->|Issue Lease| D[Specialist Agent]
        D -->|Tool Call + Lease| E[Gateway / Broker]
        E -->|Verify Signature| C
        E -->|Execute| F[Tool]
        D -->|Task Complete| G[MRCP]
        G -->|Invalidate| B
        B -->|Purge Signature| C
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(taskID, capabilities, parentLeaseID) -> SignedLease`
    * `hlml.VerifyLease(signedLease) -> Boolean`
    * `hlml.RevokeLease(taskID) -> Status`
* **Data Storage/State:**
    * **Lease Registry**: Ephemeral, hardware-protected storage for active lease fingerprints.
    * **Task-Lease Mapping**: In-memory mapping of Mission-Root tasks to their issued HLML tokens.

## 5. Alternatives Considered
* **JWT-based Time-Limited Tokens**: Rejected because time is a poor proxy for task lifecycle. Tasks can finish early or run late. Hardware-locked revocation is more deterministic.
* **OS-Level Sudo/RBAC**: Rejected because it lacks "Agentic Intent" awareness and cannot be easily sharded across distributed mesh nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lease issuance requires hardware-attested mission-root lineage (RMRA).
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time visualization of active and expired leases.

## 7. Evolutionary Changelog
* **2026-04-04:** Initial Document Creation.

### Update: [2026-04-04] (Session 2) - Resolving Mailbox Lock Contention
**Context**: Large-scale Agent Teams are experiencing 5s+ coordination stalls due to synchronous mailbox locks.
**Architecture Adjustment**:
* Integrating the **Lock-Free Teammate Coordination (LFTC)** module in Section 4.
* Transitioning the Lease Registry to a sharded, CRDT-native architecture to support non-blocking task claiming.
**Performance Impact**: Reduces coordination latency by an estimated 80% in swarms with 10+ teammates.
