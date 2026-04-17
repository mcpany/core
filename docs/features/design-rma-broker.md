# Design Doc: Recursive Mesh Attestation (RMA) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In a distributed agent mesh, tool calls often traverse intermediate nodes (relays) to reach specialized hardware (e.g., Laptop -> Bastion -> GPU Workstation). Standard point-to-point encryption protects data in transit but doesn't prevent a malicious relay from spoofing the "Origin Identity" or modifying the intent before it reaches the final node.

The Recursive Mesh Attestation (RMA) Broker ensures that identity provenance and mission-bound authority are maintained through infinite hops by utilizing hardware-attested lineage tokens.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate multi-hop, hardware-attested inter-agent communication.
    * Maintain a non-repudiable chain of custody for agent intents across relays.
    * Neutralize "Relay Trust Gaps" by enforcing end-to-end mission validation.
    * Support sub-millisecond fast-path resumption for high-frequency mesh calls.
* **Non-Goals:**
    * Replacing existing AMT (Attested Mesh Tunneling) for simple 1-hop cases; it evolves AMT.
    * Providing general-purpose P2P file sharing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Swarm Developer
* **Primary Goal:** Execute a secure reasoning step on a remote node through an intermediate cloud relay without exposing the raw context to the relay.
* **The Happy Path (Tasks):**
    1. Origin Agent (Node A) initiates a call to specialized tool on Node C.
    2. RMA Broker on Node A wraps the request in a hardware-attested mission token and sends it to Node B.
    3. Node B (Relay) adds its own attestation fragment to the lineage but cannot modify Node A's core intent.
    4. Node C receives the multi-hop token and verifies the recursive chain back to Node A's TPM signature.
    5. Node C executes the tool and returns the result through the established RMA tunnel.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        subgraph Node A (Origin)
            A[Agent] --> B[RMA Broker]
        end
        subgraph Node B (Relay)
            C[RMA Broker]
        end
        subgraph Node C (Target)
            D[RMA Broker] --> E[GPU Tool]
        end
        B -->|Attested Hop 1| C
        C -->|Attested Hop 2| D
    ```
* **APIs / Interfaces:**
    * `rma.InitializeRecursiveTunnel(targetID, relayChain, missionRoot) -> TunnelID`
    * `rma.AttestRelay(tunnelID, relayID) -> LineageToken`
    * `rma.ExecuteEndToEnd(tunnelID, payload) -> Result`
* **Data Storage/State:**
    * **Lineage Cache:** Local store of verified hardware-chain fragments to minimize re-verification latency.

## 5. Alternatives Considered
* **Mutual TLS (mTLS) per hop:** Rejected because it only proves the immediate sender's identity, not the original mission root's authority through multiple hops.
* **VPN Tunnels:** Insufficient "Agentic Awareness" for per-call mission-binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory origin-locked handshakes (SOP) are enforced at every hop.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" UI for visualizing multi-hop lineages.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
