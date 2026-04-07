# Design Doc: Atomic Swarm Transaction (AST) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms transition from simple tool-calling to distributed, high-density meshes, the migration of state between physical nodes and execution environments has become a critical bottleneck. Currently, state handoffs (especially in distributed memory enclaves like DME) are non-atomic, leading to "Cognitive Split-Brain" where different parts of the swarm operate on inconsistent state fragments during a migration event.

The AST Manager solves this by introducing transactional atomicity to the agent coordination bus. It ensures that complex state migrations across mesh nodes are either fully committed or safely rolled back, maintaining a consistent cognitive worldview for the entire swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a 2-Phase Commit (2PC) protocol for memory enclave state migrations.
    * Provide hardware-attested commit tokens to ensure migration integrity.
    * Neutralize "Cognitive Split-Brain" scenarios during node failover or re-sharding.
    * Support rollback to hardware-signed checkpoints upon migration failure.
* **Non-Goals:**
    * Replacing the underlying storage (Shared KV Store/Blackboard).
    * Managing the execution of the agent reasoning loop itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Migrate a high-trust reasoning state from an Edge node to a Core enclave without risking intent drift.
* **The Happy Path (Tasks):**
    1. The AST Manager initiates a `Prepare` phase, locking the source memory enclave.
    2. The target node generates a hardware-attested `Readiness Proof`.
    3. The source node streams the encrypted state fragments to the target.
    4. The AST Manager verifies the `Migration Hash` against the source.
    5. The `Commit` signal is broadcast, atomically updating the mesh registry to point to the new node.
    6. Locks on the source are released and fragments are purged.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant SM as AST Manager
        participant S as Source Enclave
        participant T as Target Enclave
        SM->>S: Lock State (TxID: 0x88)
        S->>SM: Lock Ack + State Hash
        SM->>T: Prepare Migration (TxID: 0x88)
        T->>SM: Readiness Proof (TPM Signed)
        SM->>S: Stream Fragments -> Target
        S->>T: [Encrypted BSH Payload]
        T->>SM: Fragment Ack + Hash Verification
        SM->>T: Commit (TxID: 0x88)
        T->>SM: Commit Success
        SM->>S: Finalize (Purge)
    ```
* **APIs / Interfaces:**
    * `POST /v1/transactions/migrate`: Initiates an atomic migration.
    * `GET /v1/transactions/{txid}/status`: Monitors migration integrity.
* **Data Storage/State:**
    * Uses a transient `Transaction Log` stored in hardware-protected memory.
    * Final state is committed to the **Universal Episodic Graph (UEG)**.

## 5. Alternatives Considered
* **Saga Pattern (Eventual Consistency)**: Rejected because "eventual consistency" in reasoning state leads to hallucination-driven cascades that are impossible to recover from.
* **Synchronous Global Locks**: Rejected due to the "Mailbox Lock" performance ceiling identified in Claude Code meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All migration proofs must be TPM-signed. Unsigned readiness proofs trigger immediate transaction abort and quarantine of the target node.
* **Observability:** Migration latency and "Split-Brain" attempts are logged to the **Action-Chain Sovereignty Monitor (ACSM)**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
