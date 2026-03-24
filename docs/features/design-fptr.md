# Design Doc: Fast-Path Teammate Resume (FPTR)
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), the overhead of hardware-bound mission re-attestation is becoming a critical performance bottleneck. Currently, every time a teammate rotates or a new task is claimed, a 200ms+ delay is incurred for full TPM-based verification. In high-density teams (10+ agents), this leads to significant "Coordination Stall."

Fast-Path Teammate Resume (FPTR) evolves the identity model by introducing mesh-resident, pipe-bound trust leases. This allows horizontal teammates to resume validated mission contexts in sub-10ms without sacrificing Zero-Trust integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce teammate rotation latency from 200ms+ to <10ms.
    * Implement "Pipe-Bound Trust" where leases are cryptographically pinned to the local Docker-bound named pipe.
    * Maintain hardware-attested lineage across all fast-path resumptions.
    * Provide automatic revocation of mesh-resident leases upon mission-root termination.
* **Non-Goals:**
    * Bypassing initial hardware attestation (first-time boot still requires full TPM signature).
    * Supporting remote (TCP/UDP) fast-path (restricted to local isolated transport).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Enable 12 parallel Claude Code teammates to claim and execute small tasks from a shared list with near-zero coordination latency.
* **The Happy Path (Tasks):**
    1. The Lead Agent performs a full hardware-attested handshake to establish the mission root.
    2. FPTR issues a "Mesh-Resident Lease" bound to the local named-pipe transport.
    3. Teammate A claims Task 1 and provides its lease token.
    4. FPTR validates the token against the local pipe-ID in <5ms.
    5. Teammate A executes Task 1 and rotates out.
    6. Teammate B claims Task 2 using the same lease-bound identity, resuming in <5ms.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[Teammate] -->|Lease Token + Pipe ID| FPTR[FPTR Provider]
        FPTR -->|Verify Local Cache| Cache[(Mesh-Resident Cache)]
        Cache -->|Success| Auth[Authenticated Session]
        Cache -->|Miss/Expire| TPM[TPM Re-Attestation]
    ```
* **APIs / Interfaces:**
    * `fptr.IssueLease(missionID, pipeID) -> LeaseToken`
    * `fptr.ResumeSession(token) -> SessionContext`
* **Data Storage/State:**
    * **Mesh-Resident Identity Cache:** In-memory, kernel-locked cache of active trust leases.

## 5. Alternatives Considered
* **Persistent WebSockets:** Rejected because long-lived connections are prone to "Team Ghosting" and resource leaks in large swarms.
* **Shared Memory Auth:** Rejected due to the complexity of secure memory mapping across heterogeneous framework containers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** FPTR relies on the absolute isolation of the named-pipe transport. If the pipe is compromised, the lease is invalidated.
* **Observability:** Integrated with the "Teammate Rotation Dashboard" to visualize latency gains.

## 7. Evolutionary Changelog
* **2026-06-30:** Initial Document Creation.
