# Design Doc: Fast-Path Identity Resumption (FPIR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Attested Mesh Tunneling (AMT) and mandatory P2P encryption, inter-node agent coordination is facing significant latency overhead. Every remote tool call currently requires a full hardware-attested handshake, which can take 100ms+, causing "Cognitive Stall" in high-frequency swarm operations.

The Fast-Path Identity Resumption (FPIR) service is designed to broker time-bound, hardware-attested trust leases (Mesh Tickets) that allow nodes to resume secure sessions without the prohibitive latency of repeated full hardware signatures.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce inter-node handshake latency to <5ms for resumed sessions.
    *   Issue hardware-locked "Mesh Tickets" that are mission-bound.
    *   Maintain Zero-Trust security by requiring re-attestation after ticket expiration.
*   **Non-Goals:**
    *   Replacing initial hardware attestation (FPIR requires a valid initial TPM handshake).
    *   Managing long-term identity storage (handled by SMI Relay).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Swarm Specialist
*   **Primary Goal:** Perform 50 consecutive remote tool calls across nodes in under 1 second.
*   **The Happy Path (Tasks):**
    1.  Agent on Node A performs an initial AMT handshake with Node B (TPM-signed).
    2.  FPIR Broker on Node B issues a 300-second "Mesh Ticket" to Node A.
    3.  Agent on Node A performs 49 subsequent tool calls using the Mesh Ticket.
    4.  FPIR on Node B validates the ticket in kernel-memory (sub-millisecond).
    5.  Tool calls execute at near-native speed.
    6.  Ticket expires; Node A automatically performs a background re-attestation to mint a new ticket.

## 4. Design & Architecture
*   **System Flow:**
    [Agent A] --- (Mesh Ticket) ---> [FPIR Validator (Node B)] ---> [Local Tool]
                                        |
                                [TPM Session Cache]

*   **APIs / Interfaces:**
    *   `POST /v1/fpir/mint`: Exchange a valid hardware-attestation for a Mesh Ticket.
    *   `POST /v1/fpir/validate`: Verify a Mesh Ticket against the session cache.
*   **Data Storage/State:**
    *   **Secure Session Cache:** Kernel-bound, non-exportable memory tracking active tickets.

## 5. Alternatives Considered
*   **Persistent Mutual TLS (mTLS):** Rejected because mTLS sessions don't natively carry agent-mission context and are difficult to rotate at sub-minute intervals without performance impact.
*   **Standard JWTs:** Rejected because standard JWTs are susceptible to exfiltration. Mesh Tickets are cryptographically bound to the node's hardware fingerprint.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Tickets are mission-bound; a ticket issued for "Mission Alpha" cannot be used to execute tools for "Mission Beta."
*   **Observability:** Monitor "Fast-Path vs. Full-Attestation" ratios via the UI "FPIR Lease Monitor."

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
