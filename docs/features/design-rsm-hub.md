# Design Doc: Reactive Shard Migration (RSM) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As multi-agent swarms increasingly operate across distributed edge and cloud environments, the "Tunneling Overhead" introduced by mandatory peer-to-peer security (e.g., Sovereign Node Tunneling) has become a primary performance bottleneck. "Mean Time to Coordinate" (MTTC) often exceeds the latency of the actual reasoning step, leading to swarm inefficiency.

The Reactive Shard Migration (RSM) Hub is an evolution of the Dynamic Mesh Resilience (DMR) architecture. It enables the proactive and dynamic relocation of sharded context fragments to physical nodes that minimize the coordination latency for the active mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time MTTC monitoring for all mesh nodes.
    * Provide proactive state migration triggers based on intent-prediction and access patterns.
    * Neutralize P2P tunneling latency by ensuring frequently-accessed shards are co-located with the primary reasoning engine.
    * Maintain hardware-attested continuity during shard movement.
* **Non-Goals:**
    * Replacing Sovereign Node Tunneling (RSM optimizes the tunnel usage, it doesn't remove the security requirement).
    * General-purpose data replication (focus is on active mission context shards).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Orchestrator
* **Primary Goal:** Minimize latency for a high-frequency code-review task distributed across 3 geographical regions.
* **The Happy Path (Tasks):**
    1. A mission-root is established on an edge device in New York.
    2. A specialist subagent in a London cloud-node is assigned a heavy code-analysis task.
    3. Sovereign Node Tunneling (SNT) establishes a secure but high-latency bridge between NYC and London.
    4. RSM Hub detects the high MTTC (300ms+) and frequent mailbox access.
    5. RSM proactively migrates the NY mission-root shard to the London node (or a closer high-performance peer).
    6. Coordination latency drops to sub-10ms, significantly accelerating the code-review process.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Monitor[MTTC Monitor] -->|Latency Metrics| Arbiter[RSM Arbiter]
        Intent[Intent Engine] -->|Access Prediction| Arbiter
        Arbiter -->|Migration Trigger| Migrator[Shard Migrator]
        Migrator -->|Transfer| ESB[Entangled State Broker]
        ESB -->|New Affinity| NewNode[Optimal Physical Node]
    ```
* **APIs / Interfaces:**
    * `GET /rsm/topology`: View current shard distribution and MTTC heatmap.
    * `POST /rsm/migrate/proactive`: Force a proactive migration based on custom rules.
    * `Middleware: RSM_Latency_Interceptor`: Measures coordination overhead in real-time.
* **Data Storage/State:**
    * Migration metadata is stored in a mesh-synchronized "RSM Ledger" using CRDTs for consensus.

## 5. Alternatives Considered
* **Static Placement Rules:** Rejected because agent missions are too dynamic for manual region-pinning.
* **Global Anycast State:** Rejected due to the extreme complexity of maintaining hardware-attested consistency across an anycast mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Migration is gated by "Auth-at-the-Pipe" and requires destination-node hardware re-attestation before the shard is "unlocked" for local use.
* **Observability:** Integrated with the "Global Agent Activity Map" to visualize state movement across the mesh.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
