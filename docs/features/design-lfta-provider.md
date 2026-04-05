# Design Doc: Attestation Lease (LFTA v2.5) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the expansion of AI swarms into high-frequency tool-calling environments, the 100ms+ overhead of full hardware-attested handshakes (TPM signatures) per call has become a critical performance bottleneck, causing "Attestation Jitter" or "Cognitive Stuttering." Agents executing rapid sequences of local or remote tools are stalled by the security protocol itself.

The Attestation Lease (LFTA v2.5) Provider introduces a time-bound, hardware-locked trust resumption mechanism. It allows an agent to perform a full handshake once and receive a "Trust Lease" (session ticket) that can be used for sub-millisecond validation of subsequent tool calls within a verified mission scope.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce per-call attestation latency from 100ms+ to <1ms.
    * Maintain hardware-bound (TPM) security strength using cryptographically signed session tickets.
    * Enforce mission-bound and role-bound scope restrictions on all leases.
    * Support "Lightweight Resumption" across sharded teammate meshes.
* **Non-Goals:**
    * Eliminating hardware attestation entirely; it only optimizes the frequency of full handshakes.
    * Providing long-term persistent credentials; leases are strictly time-bound and mission-bound.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency DevOps Swarm
* **Primary Goal:** Execute a sequence of 50 shell commands across 3 nodes in under 2 seconds without compromising Zero-Trust integrity.
* **The Happy Path (Tasks):**
    1. The supervisor agent initiates the mission and performs a full TPM-bound handshake with the MCP Any gateway.
    2. The LFTA Provider issues a 5-minute "Attestation Lease" bound to the Mission ID and the agent's Hardware ID.
    3. The agent begins executing shell commands.
    4. For each command, the gateway validates the "Attestation Lease" ticket in memory (sub-millisecond) instead of triggering a new TPM signature.
    5. The mission completes. The agent discards the lease, or it expires automatically.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant Gateway
        participant LFTA_Provider
        participant TPM

        Agent->>Gateway: Initial Tool Call (Full Auth)
        Gateway->>TPM: Request Hardware Signature
        TPM-->>Gateway: Hardware Attestation
        Gateway->>LFTA_Provider: Mint Lease (Mission-Bound)
        LFTA_Provider-->>Agent: Lease Ticket (LFTA v2.5)

        loop Rapid Tool Sequence
            Agent->>Gateway: Tool Call + Lease Ticket
            Gateway->>LFTA_Provider: Fast-Path Verify Ticket
            LFTA_Provider-->>Gateway: Valid (Sub-ms)
            Gateway->>Tool: Execute
        end
    ```
* **APIs / Interfaces:**
    * `lfta.MintLease(authContext, ttl) -> LeaseTicket`: Generates a hardware-bound session ticket.
    * `lfta.VerifyLease(leaseTicket, currentMission) -> Claims`: Validates the ticket against the active mission root.
    * `lfta.RevokeLease(leaseID)`: Immediately invalidates a lease upon compromise signal.
* **Data Storage/State:**
    * **Lease Registry:** High-performance, in-memory KV store for active tickets.
    * **Hardware Anchor Cache:** Local cache of verified hardware fingerprints to avoid redundant registry lookups.

## 5. Alternatives Considered
* **Persistent JWTs:** Rejected because they lack hardware-binding and are susceptible to exfiltration.
* **MFA Per-Call:** Rejected as it completely breaks autonomous agent flows due to human-in-the-loop latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets are cryptographically bound to the Mission Root. If an agent attempts to use a lease outside its authorized mission branch, the LFTA Provider triggers a full re-attestation or lockout.
* **Observability:** Leased calls are logged with a "Fast-Path" flag in the audit trail for performance analysis.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
