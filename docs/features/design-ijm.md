# Design Doc: Identity-Jitter Mitigator (IJM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The industry pivot toward mandatory hardware-attested NHI (Non-Human Identity) has hit a physical scaling bottleneck. TPM/Secure Enclave signing operations have high latency (100ms+), which becomes a performance ceiling for high-density swarms (100+ agents). This leads to "Identity Jitter," where agents fail to join the mesh or drop out during high-frequency coordination.

The Identity-Jitter Mitigator (IJM) provides a performance-optimized identity minting service that utilizes time-bound "Identity Leases" with asynchronous background refresh, decoupling the agent's reasoning speed from the TPM's physical signing speed.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce swarm formation latency for high-density Agent Teams.
    * Provide sub-millisecond identity verification via leased fast-path tokens.
    * Maintain hardware-locked security by performing asynchronous TPM re-attestation.
    * Neutralize "Identity Jitter" during mesh-wide coordination events.
* **Non-Goals:**
    * Eliminating the TPM; IJM *optimizes* how we interact with the hardware root of trust.
    * Replacing A2A authentication; it provides the high-performance tokens used *by* the A2A stack.

## 3. Critical User Journey (CUJ)
* **User Persona:** Large-Scale Swarm Architect
* **Primary Goal:** Rapidly spin up a 50-agent "Coding Factory" mesh without hitting a 5-second startup delay due to serial TPM signing.
* **The Happy Path (Tasks):**
    1. Orchestrator requests 50 agent identities.
    2. IJM Broker issues 50 "Identity Leases" (fast-path tokens) immediately.
    3. Agents join the mesh and begin reasoning in sub-100ms.
    4. In the background, the IJM Broker queues TPM signing requests to "hard-attest" the leases.
    5. As TPM operations complete, the IJM updates the Mesh Identity Registry with hard-attestation proofs.
    6. Teammates verify each other using the low-latency fast-path tokens, backed by the background attestation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[IJM Broker]
        B --> C[Issue Identity Lease]
        C --> D[Agent Join Mesh]
        B --> E[Async TPM Queue]
        E --> F[Hard Attestation]
        F --> G[Update Registry]
    ```
* **APIs / Interfaces:**
    * `ijm.RequestLease(agentMetadata) -> IdentityLease`: Quick-start token.
    * `ijm.VerifyLease(token) -> bool`: High-speed verification.
    * `ijm.RefreshAttestation(leaseID) -> HardwareProof`: Background TPM operation.
* **Data Storage/State:**
    * **Lease Store:** Fast-access in-memory cache for active identity leases.
    * **Attestation Queue:** Prioritized queue for hardware-signing operations.

## 5. Alternatives Considered
* **Parallel TPM Hardware:** Rejected due to cost and physical constraints in local execution environments. IJM provides a software-mediated performance layer for existing hardware.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases have a very short time-to-live (TTL) and are immediately revoked if the background TPM attestation fails.
* **Observability:** Status monitored via the "Hardware Trust Status Widget" and "Mesh Identity Manager."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
