# Design Doc: Semantic Shard Mirroring (SSM) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward distributed meshes, the availability and consistency of context shards (memory fragments) become critical. Currently, context is often siloed or subject to single points of failure. If a node hosting a critical intent shard goes offline, the entire swarm may suffer from "Context Amnesia."

MCP Any needs to solve this by providing a standardized, secure, and fault-tolerant mirroring protocol. The SSM Broker will manage the replication of context shards across multiple nodes, ensuring that mission-root sovereignty is maintained even during node failures, while simultaneously protecting against synchronization-time exploits.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide high-availability for context shards across a distributed agent mesh.
    * Implement real-time consistency checks using hardware-attested hashes.
    * Neutralize TOCTOU (Time-of-Check to Time-of-Use) vulnerabilities during shard synchronization.
    * Support seamless failover for specialist agents accessing mirrored state.
* **Non-Goals:**
    * Building a general-purpose distributed database (this is specific to agentic context).
    * Handling persistent storage for non-agentic data.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Orchestrator
* **Primary Goal:** Ensure that a "Legal Specialist" subagent can access the verified mission-root intent even if the primary gateway node restarts.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent creates a "Legal Constraints" context shard.
    2. The SSM Broker identifies available mesh nodes and initiates mirroring.
    3. Shards are cryptographically signed and transmitted to secondary nodes.
    4. The primary node fails during a complex reasoning loop.
    5. The "Legal Specialist" subagent detects the failure and transparently switches to a mirrored shard.
    6. The SSM Broker validates the consistency of the secondary shard before allowing ingestion.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Root as Mission-Root Node
        participant Broker as SSM Broker
        participant Peer as Mesh Peer
        Root->>Broker: Push Shard Fragment
        Broker->>Broker: Generate Hardware-Attested Hash
        Broker->>Peer: Mirror Shard + Attestation
        Peer->>Broker: Ack + Signature
        Note over Root,Peer: Synchronization Window Closed
        Peer->>Broker: Heartbeat + Hash Verification
    ```
* **APIs / Interfaces:**
    * `POST /v1/ssm/mirror`: Initiates mirroring for a specific shard.
    * `GET /v1/ssm/health`: Checks consistency across mirrored nodes.
    * `RPC SyncShard`: Low-latency gRPC stream for real-time replication.
* **Data Storage/State:**
    * Shards are stored in memory-mapped buffers (memfd) for performance.
    * Consistency state is managed in a local SQLite metadata store.

## 5. Alternatives Considered
* **Vanilla Raft/Paxos:** Rejected due to high coordination overhead for transient agentic state. SSM focuses on "Semantic Consistency" rather than strict serializability.
* **Centralized Redis:** Rejected as it introduces a single point of failure and violates the "Universal Adapter" goal of local sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All mirroring traffic is encrypted via AMT (Attested Mesh Tunneling). TOCTOU is neutralized by mandatory pre-ingestion attestation of the shard hash.
* **Observability:** Mirroring latency and shard consistency scores are exposed via the System Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
