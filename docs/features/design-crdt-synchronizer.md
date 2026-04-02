# Design Doc: CRDT-Native State Synchronizer
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms transition from hierarchical, single-threaded models to horizontal "Agent Teams" (pioneered by Claude Code), the primary performance bottleneck has shifted from model reasoning latency to inter-agent coordination latency. Current "Mailbox Lock" patterns lead to "Coordination Stalls" where 3+ agents compete for write-access to a shared context shard.

The **CRDT-Native State Synchronizer** introduces Conflict-Free Replicated Data Types to the MCP Any Blackboard, enabling lock-free, parallel state mutations across disparate agent frameworks.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a lock-free synchronization layer for inter-agent coordinate messages.
    * Support multiple concurrent writers to the same context shard without coordination deadlocks.
    * Ensure "Strong Eventual Consistency" across heterogeneous frameworks (Claude, OpenClaw, AutoGen).
    * Implement shard-level hardware attestation for every CRDT mutation.
* **Non-Goals:**
    * Replacing the persistent SQLite Blackboard (CRDTs handle the "Active Coordination" layer; SQLite handles "Episodic Memory").
    * Solving model-level hallucinations (this is a transport and state integrity layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 specialist agents in a parallel refactoring task without "Mailbox Lock" timeouts.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns 5 agents and initializes a "Mission-Root Shard" in MCP Any.
    2. Agents begin writing tasks and reasoning traces to the shard simultaneously.
    3. The CRDT-Native State Synchronizer merges these operations locally without requiring a global lock.
    4. Each agent sees a consistent, merged view of the swarm's progress in sub-10ms.
    5. MCP Any signs the final merged state as the "Consensus Truth" and commits it to the persistent Blackboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        AgentA[Agent A] -->|Mutation| CRDT[CRDT Synchronizer]
        AgentB[Agent B] -->|Mutation| CRDT
        CRDT -->|Local Merge| Shard[Active Context Shard]
        Shard -->|Hardware Signature| Attestation[TPM Attestation Provider]
        Attestation -->|Signed Consensus| PersistentStore[(SQLite Blackboard)]
    ```
* **APIs / Interfaces:**
    * `ApplyMutation(agent_id, op_type, payload)`
    * `GetMergedView(shard_id) -> state_tree`
    * `VerifyShardIntegrity(shard_id, signature)`
* **Data Storage/State:**
    * In-memory CRDT graphs (utilizing Yjs or Automerge patterns).
    * State fragments bound to hardware-enclave IDs.

## 5. Alternatives Considered
* **Distributed Locking (Redis/Zookeeper style):** Rejected due to the 50ms+ latency tax and "Ghost Lock" risks during agent termination.
* **Master-Agent Mediation:** Rejected because it creates a single point of failure and a "Reasoning Bottleneck."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Each CRDT operation must be cryptographically bound to a hardware-attested agent identity. "Rogue Mutation" detection is handled via the Stylometric Identity Anchoring (SIA) layer.
* **Observability:** Integrated with the "Lock-Free Coordination Debugger" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
