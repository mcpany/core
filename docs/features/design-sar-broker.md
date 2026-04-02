# Design Doc: State-Aware Routing (SAR) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms transition to sharded, multi-node meshes (e.g., OpenClaw ContextMesh v4.0), context shards are physically distributed across the network. Current routing models are "capability-blind" to state residency, leading to excessive data transfer as large context objects are moved to the execution node.

The SAR Broker optimizes swarm performance by routing tool calls to the node where the required context shard already resides, minimizing the "Coordination Tax" and reducing network-level intent leakage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a routing engine that prioritizes context residency.
    * Maintain a real-time manifest of context shard locations (Residency Map).
    * Support fallback to standard capability-based routing if the resident node is at capacity.
* **Non-Goals:**
    * Implementing the physical memory sharding itself (handled by ZCMB).
    * Global load balancing across disparate geographically separated clusters (focus is on the local/mesh swarm).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Execute a data-intensive reasoning task across 10 nodes without saturating the local network.
* **The Happy Path (Tasks):**
    1. Agent A initiates a tool call requiring Context Shard X.
    2. SAR Broker consults the Residency Map and identifies Node B as the host of Shard X.
    3. SAR Broker routes the tool call metadata to Node B.
    4. Node B executes the tool locally against Shard X and returns only the result fragment.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Request] --> SAR[SAR Broker]
        SAR --> Map{Residency Map}
        Map -->|Node B has Shard| SAR
        SAR --> NodeB[Node B Execution]
        NodeB --> Result[Result Fragment]
        Result --> Agent
    ```
* **APIs / Interfaces:**
    * `GET /mesh/residency`: Returns current shard-to-node mapping.
    * `POST /mesh/route`: Routes a tool call based on intent and shard dependency.
* **Data Storage/State:** Residency Map stored in the Mesh-Aware Blackboard with sub-millisecond TTL.

## 5. Alternatives Considered
* **Broadcast Replication**: Replicating all shards to all nodes. Rejected due to 10x memory overhead and synchronization latency.
* **Passive Cache-Routing**: Routing randomly and caching shards. Rejected due to "Cold Start" penalties in dynamic missions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Routing metadata is signed by the mission-root to prevent "Shard Probing" by rogue nodes.
* **Observability:** Visualized via the "Mesh Traffic Heatmap" showing residency-hit ratios.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
