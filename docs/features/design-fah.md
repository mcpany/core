# Design Doc: Federated Attestation Hub (FAH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms move from hierarchical, single-node architectures to decentralized, multi-node meshes, the reliance on a central "Trust Registry" becomes a bottleneck and a single point of failure. Recent updates in OpenClaw (FAQ) and the upcoming EU AI Act demand a system that can reach consensus on tool safety and agent reputation across disparate, peer-to-peer nodes.

MCP Any needs the Federated Attestation Hub (FAH) to orchestrate decentralized trust. It allows nodes to peer-attest to the safety of dynamic skills and the reliability of specialists without requiring a global coordinator.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate decentralized consensus (quorums) on tool and subagent reputation.
    * Provide hardware-attested "Trust Beacons" for peer nodes.
    * Neutralize "Attestation Deadlocks" via mission-aligned resolution policies.
    * Implement "Audit-Grade" logging for regulatory compliance (EU AI Act).
* **Non-Goals:**
    * Acting as a global central registry (FAH is inherently peer-to-peer).
    * Validating model reasoning logic (handled by RPW Enforcer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Securely delegate a high-privilege shell task to an OpenClaw specialist in a remote device mesh without a central VPC.
* **The Happy Path (Tasks):**
    1. Parent agent in Node A requests a shell tool from Specialist in Node B.
    2. Node B issues a "Trust Challenge" to Node A.
    3. Node A's FAH broadcasts a "Reputation Request" to neighboring Peer Nodes (C and D).
    4. Peers C and D return hardware-signed attestation tokens verifying Node A's mission-root lineage.
    5. Node B's FAH verifies the quorum and establishes an Attested Mesh Tunnel (AMT).
    6. Task execution proceeds with non-repudiable audit logs.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Delegate| B[Specialist Agent]
        B -->|Tool Request| C[Local FAH]
        C -->|Reputation Probe| D[Peer Nodes]
        D -->|Signed Attestation| C
        C -->|Quorum Verified| E[Execution Bridge]
        E -->|Compliance Log| F[Regulatory Vault]
    ```
* **APIs / Interfaces:**
    * `/v1/attestation/probe`: P2P endpoint for requesting peer reputation signals.
    * `/v1/attestation/quorum/resolve`: Logic for weighting peer votes based on hardware-attested proximity and historical accuracy.
* **Data Storage/State:**
    * Peer reputation scores are stored in a DHT-backed (Distributed Hash Table) local cache, cryptographically bound to Inodes.

## 5. Alternatives Considered
* **Centralized Trust Authority (CTA):** Rejected due to the "Offline-First" and "Sovereign-Mesh" requirements of OpenClaw v3.6+.
* **Static Allow-lists:** Rejected because it cannot handle dynamic skill grafting and on-demand tool discovery in multi-node environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All FAQ messages must be signed using TPM-resident keys. Any node providing "Hallucinatory Reputation" is automatically quarantined via the CSAD Hub.
* **Observability:** Quorum resolution latency is tracked in the Service Mesh Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
