# Design Doc: Fast-Path Mesh Resumption (FPMR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Sovereign Node Tunneling (SNT) in OpenClaw, inter-agent communication across physical nodes has become cryptographically secure. However, this security comes with a high latency cost (150ms+ per handshake), which severely impacts the performance of real-time tool execution in distributed meshes. Transient network disconnections frequently trigger full re-negotiations, leading to "Tunneling Fatigue."

MCP Any needs to solve this by providing a high-performance resumption mechanism that maintains absolute sovereignty while reducing the attestation tax for known teammates.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce P2P tunnel resumption latency to <10ms.
    * Utilize hardware-attested "Trust Tickets" for session continuity.
    * Ensure that trust tickets are mission-bound and non-reusable across different mission roots.
    * Neutralize "Mesh Shadowing" by requiring periodic full re-attestation.
* **Non-Goals:**
    * Replacing SNT for initial node-to-node handshakes.
    * Supporting resumption for unauthenticated or untrusted nodes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Maintain sub-millisecond tool execution across a 3-node local mesh despite transient Wi-Fi drops.
* **The Happy Path (Tasks):**
    1. Node A performs a full SNT handshake with Node B, verified by the TPM.
    2. Node A issues a hardware-attested Trust Ticket to Node B, valid for the current mission root.
    3. The connection drops briefly due to network jitter.
    4. Node B attempts to resume the tunnel by presenting the Trust Ticket.
    5. MCP Any on Node A validates the ticket's signature and mission-root binding in <2ms.
    6. The secure P2P tunnel is resumed immediately without a full 4-way handshake.

## 4. Design & Architecture
* **System Flow:**
    [Node A] --(Full SNT Handshake + TPM Auth)--> [Node B]
    [Node A] --(Issue Mission-Bound Trust Ticket)--> [Node B]
    ... [Network Drop] ...
    [Node B] --(Resumption Request + Trust Ticket)--> [Node A]
    [Node A] --(Validate Ticket + Mission Binding)--> [Node B]
    [Node A] <==(Secure Fast-Path Tunnel)==> [Node B]

* **APIs / Interfaces:**
    * `POST /mesh/resumption/ticket`: Issue a new trust ticket for a specific peer and mission.
    * `POST /mesh/resumption/resume`: Use a ticket to re-establish an encrypted tunnel.
* **Data Storage/State:**
    * Trust Tickets are stored in kernel-bound memory (HLES) and are never persisted to disk.
    * State is synchronized with the Atomic Shard Lock-Manager (ASLM) to prevent race conditions during resumption.

## 5. Alternatives Considered
* **Persistent Long-Lived Sessions:** Rejected due to increased risk of identity hijacking if a node is compromised.
* **Standard mTLS Session Resumption:** Rejected as it lacks hardware-attested mission-root binding required for zero-trust agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets are HMAC-signed with keys residing in the Secure Enclave. Any mismatch in mission-root or session-ID triggers immediate revocation.
* **Observability:** Track resumption success rates and latency gains in the "Fast-Path Mesh Resumption (FPMR) Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
