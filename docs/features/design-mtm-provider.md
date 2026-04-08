# Design Doc: Mesh Topology Masking (MTM) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward distributed P2P meshes (e.g., OpenClaw v3.6.1 SNT), the physical and logical topology of the mesh itself becomes an attack surface. Today's market sync revealed a "Tunnel-Mapping" vulnerability where malicious subagents can infer the network topology, node distance, and potentially identify high-value targets (like hardware-enclaves or secure databases) by performing fine-grained latency analysis on inter-node coordination messages.

The Mesh Topology Masking (MTM) Provider is required to neutralize these side-channel attacks by decorrelating execution latency from network topology.

## 2. Goals & Non-Goals
* **Goals:**
    * Inject hardware-attested, intent-aware jitter into all inter-node Coordination fragments.
    * Normalize response times across the mesh to prevent node distance inference.
    * Provide a "Topology-Agnostic" view of the mesh to connected agents.
    * Integrate with the Attested Mesh Tunneling (AMT) broker to mask tunnel initiation signatures.
* **Non-Goals:**
    * Masking low-level TCP/UDP traffic patterns outside the MCP Any bus.
    * Replacing standard encryption (which MTM complements).
    * Impacting the performance of latency-critical local tool calls (jitter is applied only to inter-node mesh traffic).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Operator
* **Primary Goal:** Prevent a compromised specialist subagent from mapping the location of a secure hardware-vault node within a 10-node mesh.
* **The Happy Path (Tasks):**
    1. The operator enables MTM across the mesh via the Governance Center.
    2. A subagent on Node A sends a coordination message to the Vault on Node J.
    3. The MTM Provider on Node A calculates the current "Mesh Noise Profile" based on the mission-root's sensitivity.
    4. The MTM Provider injects a calculated jitter (e.g., 5ms - 50ms) into the response path.
    5. The subagent attempts to measure the RTT (Round Trip Time) to identify Node J's proximity.
    6. Due to the intentional jitter, the RTT remains non-deterministic and decorrelated from the physical hop count.
    7. The subagent's topology-mapping attempt fails to yield a confident mesh graph.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Request] --> B[AMT Broker]
        B --> C{MTM Middleware}
        C -->|Intent Analysis| D[Jitter Engine]
        D -->|Hardware-Attested Entropy| E[Coordination Transport]
        E --> F[Remote Node]
    ```
* **APIs / Interfaces:**
    * `mtm.MaskLatency(requestID, sensitivityLevel)`: Injects jitter into the response pipeline.
    * `mtm.GetNoiseProfile(missionToken)`: Retrieves the current jitter configuration for a mission branch.
* **Data Storage/State:**
    * **Jitter Entropy Store:** Hardware-bound seed for generating non-deterministic timing profiles.
    * **Topology Baseline:** Cached RTT measurements used to calculate the required normalization jitter.

## 5. Alternatives Considered
* **Static Jitter (Constant Delay):** Rejected because it introduces a fixed performance penalty that is easy for RL-optimized agents to "filter out" via statistical averaging.
* **Network-Level Tor/Onion Routing:** Rejected due to the extreme latency tax and complexity of managing agent-bound identities through multiple hops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MTM itself must be hardware-attested to ensure the jitter engine hasn't been disabled by an attacker.
* **Observability:** The "Mesh Topology & Latency Map" in the UI will show the "Masked" vs. "Real" latency to administrators while hiding the "Real" metrics from agent-facing logs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
