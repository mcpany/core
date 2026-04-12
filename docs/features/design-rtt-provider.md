# Design Doc: Recursive Trust Provider (RTP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In distributed multi-node agent meshes, agents frequently delegate tasks across multiple hops (e.g., Node A -> Node B -> Node C). Existing attestation models suffer from "Handshake Fatigue," where each hop requires a full, high-latency hardware attestation. Recursive Trust Tickets (RTT), introduced in OpenClaw v3.6.2, allow trust to be bundled and propagated through the mesh transport layer.

The Recursive Trust Provider (RTP) in MCP Any will act as the authoritative "Trust Relay," issuing and validating RTTs to ensure multi-hop trust continuity with minimal latency overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested Recursive Trust Tickets that bundle agent lineage and mission authority.
    * Facilitate sub-10ms trust verification for multi-hop delegations.
    * Neutralize "Handshake Fatigue" in high-density distributed swarms.
    * Ensure that trust strength does not degrade across mesh hops.
* **Non-Goals:**
    * Replacing the primary hardware root of trust (TPM).
    * Managing non-agentic network encryption (handled by AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Swarm Orchestrator
* **Primary Goal:** Execute a secure tool chain across three different physical devices without repeated user approval or high-latency handshakes.
* **The Happy Path (Tasks):**
    1. Primary Agent on Node A initiates a mission and obtains a root RTT from the RTP.
    2. Agent on Node A delegates a sub-task to Node B, passing the RTT.
    3. Node B's RTP validates the RTT and issues a derived RTT for Node C.
    4. Node C's RTP verifies the derived RTT, confirming the lineage back to Node A's hardware root.
    5. Node C executes the tool securely, anchored to the original mission intent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant NodeA as Node A (Root)
        participant NodeB as Node B (Relay)
        participant NodeC as Node C (Leaf)
        NodeA->>NodeA: Generate Root RTT (TPM Signed)
        NodeA->>NodeB: Delegate Task + RTT
        NodeB->>NodeB: Verify RTT & Issue Derived RTT
        NodeB->>NodeC: Delegate Task + Derived RTT
        NodeC->>NodeC: Verify Lineage & Execute
    ```
* **APIs / Interfaces:**
    * `rtp.IssueTicket(parentTicket, missionScope) -> RTT`: Generates a new or derived trust ticket.
    * `rtp.ValidateTicket(ticket) -> Claims`: Verifies the cryptographic integrity and lineage of a ticket.
* **Data Storage/State:**
    * **Ticket Registry:** Short-lived, in-memory cache of active RTT fingerprints to prevent replay attacks.

## 5. Alternatives Considered
* **Per-Hop Full Attestation:** Rejected due to prohibitive latency (500ms+ per hop).
* **Static Mesh API Keys:** Rejected due to lack of mission-bound granularity and high risk of exfiltration.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RTTs are time-bound and cryptographically restricted to the specific mission branch. Revocation is propagated via the ARL Provider.
* **Observability:** Integrated with the "Context Chain Inspector" for visual audit of trust propagation.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
