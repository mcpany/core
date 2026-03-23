# Design Doc: Teammate Mailbox Shard Arbiter (TMSA)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
As horizontal agent swarms (e.g., Claude Code Agent Teams) scale beyond 10+ teammates, the "Mailbox Lock" bottleneck becomes the primary latency driver. In a Conflict-Free Replicated Data Type (CRDT) environment, multiple teammates may attempt to claim or mutate task state simultaneously, leading to coordination stalls and high CPU usage due to optimistic concurrency failures.

The Teammate Mailbox Shard Arbiter (TMSA) provides a kernel-level coordination layer that manages concurrent write-access to granular mailbox shards. By utilizing a "Fast-Path Lease" model, the TMSA allows teammates to obtain exclusive, time-bound write-access to a specific shard without global state locks.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate coordination stalls in high-density horizontal meshes (20+ teammates).
    * Manage concurrent write-access to CRDT-based mailbox shards via lightweight leases.
    * Ensure atomic task claiming across heterogeneous framework boundaries.
* **Non-Goals:**
    * Replacing CRDTs as the underlying state synchronization mechanism.
    * Managing the content of teammate messages (handled by Mailbox Integrity Middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Coordinate a team of 25 agents to process a complex code migration without mailbox lock contention.
* **The Happy Path (Tasks):**
    1. The Lead Agent partitions the migration task into 25 shards in the shared mailbox.
    2. Specialist Agent A requests a "Shard Lease" for `shard_001` from the TMSA.
    3. The TMSA grants the lease, pinning Specialist Agent A's hardware-bound identity to the shard.
    4. Specialist Agent A processes the task and commits the CRDT delta.
    5. Upon completion, the TMSA releases the lease, and the shard becomes available for subsequent refinement.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Teammate->>TMSA: Request Shard Lease (Shard ID, Identity Token)
        TMSA->>TMSA: Validate Hardware Identity
        TMSA->>Shard Store: Check Lease Status
        Shard Store-->>TMSA: Shard Available
        TMSA-->>Teammate: Lease Granted (TTL: 500ms)
        Teammate->>Shard Store: Commit CRDT Delta
        Teammate->>TMSA: Release Lease
    ```
* **APIs / Interfaces:**
    * `POST /v1/mailbox/lease/acquire`: Requests a time-bound lease for a specific shard.
    * `POST /v1/mailbox/lease/release`: Explicitly releases a lease.
    * `GET /v1/mailbox/lease/status`: Returns current lease ownership across the mesh.
* **Data Storage/State:**
    * Lease registry is held in high-speed, kernel-bound shared memory.
    * Persisted lease logs are stored in the Namespace-Locked Registry for audit.

## 5. Alternatives Considered
* **Global Optimistic Locking:** Rejected as it leads to "Collision Storms" as teammate density increases.
* **Centralized Redis Locking:** Rejected due to the latency overhead of network round-trips for local inter-agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lease acquisition requires hardware-attested identity tokens bound to the mission-root.
* **Observability:** "Lease Contention Metrics" track shards with high request density, triggering automated sharding refinement.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
