# Design Doc: Fast-Path Identity Resumption (FPIR) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of Sovereign Node Tunneling (SNT) in OpenClaw and mandatory P2P encryption in distributed Agent Teams has introduced a significant performance tax. Full hardware-bound handshakes for every cross-device tool call add 100ms+ of latency, leading to "Cognitive Stall" in complex reasoning loops.

The FPIR Broker provides a high-performance trust-shuttle mechanism. It issues time-bound, hardware-attested "Trust Tickets" that allow agents to resume secure tunnels and re-verify identities across distributed nodes with sub-millisecond overhead, maintaining absolute sovereignty without the latency tax.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-node handshake latency from 100ms+ to <1ms.
    * Broker hardware-attested (TPM/Secure Enclave) trust tickets for P2P tunnel resumption.
    * Maintain "Lineage Continuity" across device boundaries.
    * Support automated ticket revocation via the ARL Provider.
* **Non-Goals:**
    * Replacing primary hardware handshakes for initial mission-root establishment.
    * Managing non-P2P (cloud-to-local) authentication.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed LLM Swarm Orchestrator
* **Primary Goal:** Execute a sequence of 50+ tool calls across 3 different devices (Laptop, Home Server, Edge Gateway) with minimal latency.
* **The Happy Path (Tasks):**
    1. The agent establishes an initial hardware-attested handshake with the FPIR Broker on Device A.
    2. The Broker issues a "Fast-Path Trust Ticket" cryptographically bound to the mission root and Device A's TPM.
    3. The agent migrates a sub-task to Device B and presents the ticket.
    4. Device B's FPIR listener validates the ticket signature locally against the Broker's public key in <1ms.
    5. The secure P2P tunnel is resumed instantly without a full re-handshake.
    6. Tool execution proceeds at native speeds.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Start] --> B[Full Hardware Handshake]
        B --> C[FPIR Ticket Issuance]
        C --> D[Distribute Ticket to Teammates]
        D --> E{Cross-Node Call?}
        E -- Yes --> F[Present FPIR Ticket]
        F --> G[Local Signature Validation]
        G --> H[Instant Tunnel Resumption]
        E -- No --> I[Standard Local Execution]
    ```
* **APIs / Interfaces:**
    * `IssueTrustTicket(handshakeProof, ttl) -> TrustTicket`
    * `VerifyTrustTicket(ticket) -> Claims`
    * `RevokeTicket(ticketID) error`
* **Data Storage/State:**
    * Tickets are stored in kernel-bound protected memory on each node, synchronized via the AMS (Asynchronous Mailbox Sharding) layer to ensure consistent revocation.

## 5. Alternatives Considered
* **Persistent mTLS Connections**: Rejected due to the high resource overhead of maintaining thousands of open sockets in large meshes.
* **Session-less UDP**: Rejected as it lacks the required hardware-attested security posture for high-privilege tool calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets are mission-bound and hardware-locked; they cannot be re-used by rogue subagents on the same device if the process environment is different.
* **Observability:** Latency savings from FPIR usage are surfaced in the "Binary Handoff Performance Monitor" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
