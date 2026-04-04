# Design Doc: Autonomous State Migration (ASM) Controller
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI Agent Teams (e.g., Claude Code, OpenClaw swarms) scale across distributed meshes, the failure of a single node or the revocation of a hardware attestation can lead to "Cognitive Stall" or total state loss for the mission. Currently, agent state is often pinned to the local node where the subagent was spawned.

The Autonomous State Migration (ASM) Controller is needed to ensure fail-operational resilience by automatically re-sharding and migrating mission-critical state fragments between physical nodes when a node becomes unhealthy or its security posture drifts.

## 2. Goals & Non-Goals
* **Goals:**
    * Proactively detect node failures or attestation breaches in the agent mesh.
    * Automatically migrate Shared KV Store (Blackboard) shards and Context Shards to healthy nodes.
    * Maintain mission-root consistency during migration using "Checkpoint-and-Resume" semantics.
    * Integrate with the Dynamic Mesh Resilience (DMR) Hub for coordinated re-sharding.
* **Non-Goals:**
    * Migrating active process execution (CPU/RAM state); it focuses on persisted cognitive state.
    * Handling load balancing for non-agent traffic.
    * Managing persistent storage redundancy (e.g., RAID); it operates at the agentic shard level.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mission-Critical Swarm Operator
* **Primary Goal:** Ensure a multi-day coding mission continues even if the primary laptop running the supervisor agent shuts down.
* **The Happy Path (Tasks):**
    1. A swarm is operating across three nodes: Laptop (Supervisor), Server A (Specialist), Server B (Auditor).
    2. The ASM Controller detects that the Laptop's battery is critical and its attestation lease is expiring.
    3. ASM triggers a "Pre-emptive Snapshot" of the Supervisor's mission state on the Laptop.
    4. ASM negotiates with Server A to become the temporary "Residency Node" for the Supervisor's state.
    5. The state shards are transferred via Attested Mesh Tunneling (AMT).
    6. Server A resumes the mission context, notifying all teammates of the new residency location.
    7. The mission continues without the Supervisor losing its chain-of-thought.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant NodeA as Source Node
        participant ASM as ASM Controller
        participant NodeB as Target Node
        participant DMR as DMR Hub

        ASM->>NodeA: Health/Attestation Check Fail
        ASM->>DMR: Initiate Re-shard
        DMR->>NodeB: Reserve Shard Capacity
        NodeA->>NodeB: Stream State Shards (via AMT)
        NodeB->>ASM: Migration Complete
        ASM->>DMR: Update Mesh Topology
    ```
* **APIs / Interfaces:**
    * `asm.MigrateShard(shardID, targetNodeID) -> MigrationToken`: Triggers a state handoff.
    * `asm.OnNodeFailure(nodeID)`: Callback for the mesh heartbeat monitor.
    * `asm.GetResidency(shardID) -> NodeID`: Locates the current physical host of a state fragment.
* **Data Storage/State:**
    * **Residency Map:** A distributed, CRDT-based table tracking the mapping of mission shards to physical nodes.
    * **Migration Log:** Append-only record of state movements for auditability.

## 5. Alternatives Considered
* **Global Distributed Database (e.g., CockroachDB):** Rejected due to the overhead of consistent global replication for every transient reasoning fragment. ASM allows for "Local Residency with Remote Failover," which is more performant for agentic reasoning.
* **Manual Checkpointing:** Rejected because it introduces "Resilience Latency" and requires the LLM to handle its own failover logic, which is error-prone.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** State fragments are encrypted with mission-bound keys that only authorized hardware enclaves can decrypt. Migration requires a valid "Migration Quorum" from the DMR Hub.
* **Observability:** Migrations are visualized in the "Mesh Resilience Dashboard" with real-time throughput metrics.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
