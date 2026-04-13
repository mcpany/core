# Design Doc: Atomic Resource Handshake (ARH) Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Attested Mesh Tunneling (AMT), distributed agent swarms have achieved high levels of security but at the cost of significant coordination latency (often 200ms+ per remote tool call). This "Tunneling Overhead" is a primary bottleneck for real-time autonomous systems where sub-millisecond response is required for local-to-remote-to-local execution loops.

The Atomic Resource Handshake (ARH) Gateway addresses this by implementing a high-performance coordination broker that utilizes pre-attested session identities and hardware-bound "Trust Tickets" to reduce handshake overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Achieve sub-10ms latency for remote tool execution within a verified mission mesh.
    * Broker hardware-bound trust tickets that persist across mission-locked sessions.
    * Implement pre-attestation of session identities to eliminate redundant TPM signatures.
    * Harmonize with OpenClaw v3.6.2 ARH standards for cross-framework interoperability.
* **Non-Goals:**
    * Reducing the latency of the underlying physical network transport (focused on coordination logic).
    * Bypassing TPM attestation (ARH optimizes the *reuse* of attested trust).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Agent Mesh Developer
* **Primary Goal:** Execute 100+ remote tool calls across a distributed swarm within a 1-minute window without hitting TPM rate limits or incurring prohibitive latency.
* **The Happy Path (Tasks):**
    1. The agent performs an initial full hardware handshake with the remote node via the AMT Broker.
    2. The ARH Gateway issues a mission-bound, hardware-locked "Trust Ticket" cached in secure memory.
    3. For subsequent tool calls, the agent provides the Trust Ticket instead of a fresh TPM signature.
    4. The remote ARH Gateway validates the ticket in <1ms and authorizes the tool execution.
    5. The mission completes, and the ARH Gateway automatically invalidates all associated Trust Tickets.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant ARH_Local
        participant ARH_Remote
        participant TPM

        Note over Agent, TPM: Initial Handshake
        Agent->>ARH_Local: Request Remote Tool
        ARH_Local->>TPM: Sign Mission Identity
        TPM-->>ARH_Local: Signature
        ARH_Local->>ARH_Remote: Full Attested Handshake
        ARH_Remote-->>ARH_Local: Mission Trust Ticket

        Note over Agent, TPM: Fast-Path Execution
        Agent->>ARH_Local: Request Remote Tool (with Ticket)
        ARH_Local->>ARH_Remote: Token-Bound Request
        ARH_Remote->>ARH_Remote: Validate Ticket (<1ms)
        ARH_Remote-->>ARH_Local: Authorize & Return Result
    ```
* **APIs / Interfaces:**
    * `arh.MintTicket(missionID, hardwareSignature) -> TrustTicket`: Issues a reusable trust token.
    * `arh.ValidateTicket(trustTicket) -> SessionStatus`: Sub-millisecond verification.
    * `arh.RevokeMission(missionID)`: Forceful invalidation of all related tickets.
* **Data Storage/State:**
    * **Secure Ticket Cache:** Kernel-bound or Enclave-local (DME) memory storing active Trust Tickets.
    * **Handshake Registry:** Maps Mission IDs to verified hardware identities.

## 5. Alternatives Considered
* **Persistent TLS Sessions:** Rejected as they don't provide mission-root granularity or hardware-binding for the specific agent session.
* **Batching Tool Calls:** Rejected because it introduces "Wait Latency" and is incompatible with sequential reasoning loops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Trust Tickets are strictly task-bound and non-exportable. Any attempt to use a ticket from an unauthorized origin triggers a hardware-level security breach.
* **Observability:** Integrated with the "ARH Trust Ticket Widget" in the UI for real-time monitoring of ticket health and latency gains.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
