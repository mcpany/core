# Design Doc: Zero-Latency Attestation Lease (ZLAL) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
High-density agent meshes (e.g., horizontal Claude Code teams) suffer from "Handshake Exhaustion." Each inter-teammate coordination event currently requires a full hardware-attested identity exchange, which can add 150-300ms of latency per hop. In deep swarms, this becomes the primary bottleneck for reasoning speed.

The Zero-Latency Attestation Lease (ZLAL) Provider solves this by introducing high-speed, hardware-bound trust leases. It uses a single hardware-heavy attestation at session start to issue short-lived, lightweight trust tokens that can be verified in sub-millisecond timeframes within the mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-teammate attestation latency to <1ms.
    * Provide hardware-bound security (TPM/Secure Enclave) for lease issuance.
    * Support automatic, background lease renewal.
    * Enable real-time revocation via the ARL Synchronizer.
* **Non-Goals:**
    * Eliminating hardware attestation entirely (it moves it to the "fast-path").
    * Managing long-term agent identities (handled by SMI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 parallel teammates on a codebase review without the 1.5s latency penalty of sequential hardware handshakes.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes the swarm and performs one full TPM-signed handshake with MCP Any.
    2. MCP Any issues a 5-minute ZLAL token to the mission root.
    3. Teammates use the ZLAL token for all inter-mailbox coordination.
    4. Verification happens instantly via local symmetric keys managed by the ZLAL Provider.
    5. The mission completes 10x faster than with per-call hardware signatures.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        AgentA[Agent A] -->|Hardware Attest| ZLAL[ZLAL Provider]
        ZLAL -->|Issue Lease| LeaseToken[Lease Token]
        LeaseToken --> AgentA
        AgentA -->|Request + Token| AgentB[Agent B]
        AgentB -->|Fast Verify| ZLAL
        ZLAL -->|Success| AgentB
    ```
* **APIs / Interfaces:**
    * `POST /v1/zlal/lease`: Exchange hardware signature for a fast-path lease.
    * `POST /v1/zlal/verify`: Sub-millisecond verification of a lease token.
    * `DELETE /v1/zlal/revoke`: Forcefully revoke a lease.
* **Data Storage/State:**
    * Lease keys are stored in non-pageable memory within the ZLAL service.
    * Active leases are tracked in a high-speed, in-memory cache.

## 5. Alternatives Considered
* **Persistent mTLS:** Rejected due to the overhead of managing certificates for thousands of ephemeral subagents.
* **Shared Session Cookies:** Rejected as they are not hardware-bound and are vulnerable to exfiltration (LOWA/ClawJacked defense).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are mission-bound and expire automatically. They are useless if moved outside the specific host/mesh environment.
* **Observability:** Lease utilization and latency gains are tracked in the HAFP monitor.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
