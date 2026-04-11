# Design Doc: Cross-Node State Synchronization (CNSS) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In multi-node agent meshes (enabled by OpenClaw SNT and AMT), synchronizing the shared state (Blackboard) and context shards across physical devices introduces significant latency (often 200ms+). This "Cross-Node State Latency" causes "Cognitive Stall," where parallel teammates wait for state consistency before proceeding with reasoning.

The CNSS Hub addresses this by implementing an optimistic state synchronization model. It allows nodes to reason against local "Predicted Shards" while performing background hardware-attested reconciliation, reducing the coordination tax for distributed Agent Teams.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement optimistic context shard replication across mesh nodes.
    * Utilize hardware-attested (TPM) hashes for background state reconciliation.
    * Provide sub-10ms state availability for remote teammate requests.
* **Non-Goals:**
    * Will not provide global strong consistency (Eventual Consistency with Conflict Resolution).
    * Will not replace the underlying P2P tunnel (AMT).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Synchronize a 50MB context shard between a local workstation and a remote edge node in <50ms.
* **The Happy Path (Tasks):**
    1. Local agent writes to a mission-bound context shard.
    2. CNSS Hub immediately streams an optimistic fragment to the Edge Node.
    3. Edge teammate begins reasoning against the optimistic fragment.
    4. CNSS Hub completes hardware attestation in the background.
    5. CNSS Hub confirms state integrity; Edge teammate commits the reasoning step.

## 4. Design & Architecture
* **System Flow:**
    [Node A] --(Optimistic Stream)--> [CNSS Hub] --(Background Attestation)--> [Node B]
* **APIs / Interfaces:**
    * `/v1/cnss/push`: Stream optimistic state fragments.
    * `/v1/cnss/reconcile`: Trigger hardware-attested shard validation.
* **Data Storage/State:**
    * Optimistic buffers stored in RAM with `memfd_create` for zero-copy access.
    * Attested state persisted in the UEG (Universal Episodic Graph).

## 5. Alternatives Considered
* **Synchronous Replication:** Rejected due to prohibitive latency (200ms+) impacting real-time agent responsiveness.
* **Centralized State Hub:** Rejected to maintain the "Sovereign Node" principle and prevent single points of failure in the mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shards are cryptographically bound to the mission-root HMA.
* **Observability:** Integrated with the **Service Mesh Topology Monitor** for real-time visualization of sync latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
