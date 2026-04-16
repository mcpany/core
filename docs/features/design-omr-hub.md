# Design Doc: Optimistic Mesh Resumption (OMR) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The transition to horizontal, multi-node agent meshes has exposed a critical performance bottleneck: heterogeneous tunnel latency. When an agent migrates its mission context between nodes with disparate secure enclaves (e.g., Apple SEP to Intel SGX), hardware-attested handshakes introduce 150ms+ spikes. These "Attestation Taxes" cause cognitive stalls in autonomous swarms.

The Optimistic Mesh Resumption (OMR) Hub solves this by allowing agents to speculatively resume reasoning against local, hardware-locked context caches while attestation quorums complete asynchronously in the background.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Speculative Buffer" for mission-root contexts to eliminate attestation latency.
    * Maintain TPM-bound "Pre-attested Shard Caches" for frequently migrated mission fragments.
    * Automatically roll back state mutations if background attestation fails.
    * Provide a standardized interface for cross-enclave intent continuity.
* **Non-Goals:**
    * Replacing the underlying Attested Mesh Tunneling (AMT) transport.
    * Eliminating hardware attestation; it merely parallelizes it.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Sub-millisecond mission resumption when an agent moves from a mobile terminal to a workstation.
* **The Happy Path (Tasks):**
    1. An agent completes a sub-task on a Laptop and initiates migration to a high-power Workstation.
    2. The OMR Hub on the Workstation detects the incoming mission-root and checks the "Pre-attested Shard Cache."
    3. Finding a valid, mission-bound fragment, the OMR Hub grants the agent "Speculative Execution" status.
    4. The agent resumes reasoning immediately, with its tool outputs held in a probabilistic buffer.
    5. Simultaneously, the AMT Broker completes the 150ms heterogeneous handshake.
    6. Upon successful attestation, the OMR Hub commits the agent's speculative outputs to the global Blackboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Incoming Mission Token] --> B{OMR Hub Cache?}
        B -- Yes --> C[Grant Speculative Status]
        B -- No --> D[Synchronous Handshake]
        C --> E[Speculative Reasoning Loop]
        E --> F[Probabilistic Output Buffer]
        G[AMT Handshake Thread] --> H{Attested?}
        H -- Success --> I[Commit Buffer to Blackboard]
        H -- Failure --> J[Force Rollback & Disconnect]
    ```
* **APIs / Interfaces:**
    * `omr.SpeculativeResume(missionToken) -> SessionID`: Initiates an optimistic session.
    * `omr.ValidateSpeculation(sessionID, proof) -> void`: Commits speculative state.
* **Data Storage/State:**
    * **Pre-attested Shard Cache:** Encrypted, TPM-bound storage for mission-root context fragments.
    * **Speculative Log:** Time-series record of all speculative actions for rollback purposes.

## 5. Alternatives Considered
* **Persistent Mesh Tunnels:** Rejected because persistent connections drain battery on mobile terminals and increase the attack surface for mesh-shadowing.
* **Lowering Attestation Strength:** Rejected as it violates the Zero Trust security pillar.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative actions are strictly read-only for host-level resources (e.g., no filesystem writes) until attestation is confirmed.
* **Observability:** Visualized in the "Mesh Resumption Dashboard" with real-time "Time Saved" metrics.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
