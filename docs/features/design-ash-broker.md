# Design Doc: Atomic Shard Handshake (ASH) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI Agent Teams scale horizontally, the overhead of identity verification for every inter-agent message (mailbox request, task claim, state sync) has become a bottleneck, often exceeding 100ms in deep swarms. Existing hardware-attested handshakes are secure but slow for high-frequency coordination.

The Atomic Shard Handshake (ASH) Broker introduces a performance-optimized verification layer that utilizes pre-cached session tickets and hardware-bound ephemeral keys to facilitate sub-10ms mesh-wide identity verification.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-agent verification latency to <10ms.
    * Utilize hardware-bound (TPM) session tickets for trust persistence.
    * Implement a "Ticket-Granting" model for local swarm coordination.
    * Ensure compatibility with OpenClaw and Claude Code horizontal coordination patterns.
* **Non-Goals:**
    * Replacing long-haul mTLS for cross-network agent communication.
    * Implementing the consensus logic for self-healing (handled by ASH Consensus Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-frequency Subagent Teammate.
* **Primary Goal:** Send 50+ coordination messages per second to teammates without incurring attestation-latency penalties.
* **The Happy Path (Tasks):**
    1. Subagent performs an initial, deep hardware-attested handshake with the ASH Broker.
    2. ASH Broker issues a "Shard Ticket" cryptographically bound to the current mission and subagent identity.
    3. Subagent attaches the Shard Ticket to all subsequent mailbox/shard requests.
    4. Recipient agents verify the ticket locally using the ASH Broker's public key (cached).
    5. The verification happens in <2ms, allowing high-speed state synchronization.
    6. Tickets are automatically rotated or revoked upon mission-phase transition.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent A] -->|Initial Deep Auth| B[ASH Broker]
        B -->|Issue Shard Ticket| A
        A -->|Message + Ticket| C[Subagent B]
        C -->|Local Ticket Verify| D[Cache/Validator]
        D -->|Accept| C
    ```
* **APIs / Interfaces:**
    * `ash.IssueTicket(identity HardwareIdentity, missionID UUID) -> Ticket`: Issues a high-speed session ticket.
    * `ash.VerifyTicket(ticket Ticket) -> (Identity, bool)`: Performs sub-millisecond local verification.
* **Data Storage/State:**
    * **Ticket Store:** A sharded, in-memory KV store for active ticket metadata and revocation lists.

## 5. Alternatives Considered
* **Persistent mTLS:** Rejected due to handshake overhead in transient subagent lifecycles.
* **Shared Symmetric Keys:** Rejected as it lacks granular identity provenance and hardware-binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets must be short-lived and bound to specific mission phases to prevent "Ticket Replay" attacks.
* **Observability:** Integrated with the "Coordination Latency Dashboard" to track ASH performance gains.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Evolving from Fast-Path Identity Resumption (2026-06-29) to support atomic mesh-wide handshakes.
