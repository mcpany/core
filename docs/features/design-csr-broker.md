# Design Doc: Cognitive Shard Replication (CSR) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Cognitive Shard Replication (CSR) in OpenClaw, agent meshes can now maintain high availability of reasoning state across distributed nodes. However, traditional replication protocols introduce significant latency (200ms+), which causes "Cognitive Stall" in high-frequency coordination.

The CSR Broker in MCP Any provides a high-performance, hardware-attested replication layer that leverages RDMA-aware zero-copy memory transfers to synchronize context shards between mesh nodes with sub-millisecond overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a CSR-compliant replication protocol for hardware-attested context shards.
    * Achieve <1ms replication latency using zero-copy memory-mapped buffers.
    * Integrate with the DMR Hub for automated recovery of failed nodes.
    * Enforce mission-root boundary isolation during shard synchronization.
* **Non-Goals:**
    * General-purpose file or database replication.
    * Supporting nodes without hardware-attestation capabilities.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Administrator
* **Primary Goal:** Ensure that a specialist subagent's reasoning state is immediately available on a standby node if the primary node fails.
* **The Happy Path (Tasks):**
    1. A "Primary Agent" updates its internal monologue in a local context shard.
    2. The CSR Broker intercepts the write and speculatively replicates the fragment to a "Standby Node" via a zero-copy RDMA buffer.
    3. The Standby Node verifies the hardware-attested fragment signature and commits it to its local mirror.
    4. Upon Primary Node failure, the DMR Hub promotes the Standby Node.
    5. The Standby Node resumes execution instantly with the latest replicated state.

## 4. Design & Architecture
* **System Flow:**
    * [Primary Node] -> (Local Write) -> [CSR Broker (Primary)]
    * [CSR Broker (Primary)] -> (RDMA Zero-Copy) -> [CSR Broker (Standby)]
    * [CSR Broker (Standby)] -> (TPM Verify) -> [Mirrored Shard]
* **APIs / Interfaces:**
    * `rpc ReplicateShard(ShardFragment) returns (Ack)`: High-speed Protobuf interface.
    * `x-mcpany-replication-factor`: Mission-level setting for shard redundancy.
* **Data Storage/State:**
    * Utilizes the Zero-Copy Memory Broker (ZCMB) for local memory management and the Entangled State Broker (ESB) for cryptographic linkage.

## 5. Alternatives Considered
* **Asynchronous Disk Mirroring:** Rejected due to 50ms+ I/O latency.
* **Consensus-Based Log Replication (Raft):** Rejected as the primary mechanism due to the "Quorum Latency Tax" on high-frequency reasoning fragments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All replicated fragments must be cryptographically bound to the mission-root hardware session.
* **Observability:** Monitor "Replication Lag" and "Shard Synchronization Delta" across the mesh.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
