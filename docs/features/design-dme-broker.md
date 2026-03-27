# Design Doc: Distributed Memory Enclave (DME) Broker
**Status:** Draft
**Created:** 2026-07-04

## 1. Context and Scope
As agent swarms move toward zero-copy state sharing to eliminate BSH (Binary State Handoff) latency, the risk of "Memory-Mapped Escape" vulnerabilities has increased. Specialist agents from disparate frameworks sharing raw memory buffers can potentially probe or corrupt each other's execution state if isolation is only enforced at the software level.

The Distributed Memory Enclave (DME) Broker solves this by leveraging hardware-enclave (TPM/SEP) boundaries to physically isolate shared reasoning regions. It provides sub-millisecond state synchronization while ensuring that cross-framework memory access is cryptographically restricted to verified mission-root fragments.

## 2. Goals & Non-Goals
* **Goals:**
    * Physically isolate shared memory regions between agents using hardware enclaves.
    * Provide sub-100ms state resumption for rotating teammates via enclave-locked snapshots.
    * Neutralize "Enclave Timing Side-Channels" via monotonic jitter injection.
    * Ensure mission-root consistency across heterogeneous framework boundaries (OpenClaw, Claude Code).
* **Non-Goals:**
    * Replacing standard BSH for non-local or low-trust frameworks.
    * Managing general-purpose agent memory (DME is for high-frequency reasoning fragments).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Share high-frequency reasoning traces between a Claude Code supervisor and an OpenClaw specialist without risk of memory exfiltration.
* **The Happy Path (Tasks):**
    1. The DME Broker initializes a hardware-locked memory region for the mission root.
    2. Claude Code (Agent A) writes a reasoning fragment to the DME.
    3. The Broker issues a hardware-attested "Enclave Token" to OpenClaw (Agent B).
    4. Agent B mounts the DME using the token. Access is restricted to the specific task-bound shard.
    5. Agent B reads the trace and appends its response with sub-millisecond latency.
    6. (Attack Path): Agent B attempts to read Agent A's private process memory. The hardware enclave triggers a physical isolation fault and revokes Agent B's token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] -- Task Shard A --> B[DME Broker]
        C[Agent B] -- Task Shard B --> B
        B --> D{Hardware Enclave}
        D -- Isolation Fault? --> E[Capability Revocation]
        D -- Verified? --> F[Zero-Copy Sync]
        G[Jitter Provider] --> B
    ```
* **APIs / Interfaces:**
    * `CreateEnclave(ctx, missionID) (EnclaveID, error)`
    * `GrantShardAccess(ctx, enclaveID, agentID, shardRange) (AccessToken, error)`
    * `SyncEnclaveState(ctx, accessToken, delta) error`
* **Data Storage/State:** State is stored in memory-mapped buffers (`memfd_create`) within the TEE (Trusted Execution Environment), with metadata anchored to the mission-root manifest.

## 5. Alternatives Considered
* **Software-Only Memory Mapping:** Rejected due to the high risk of sandbox escapes in multi-tenant environments.
* **Encrypted BSH:** Rejected for local swarms due to the 150ms+ serialization/decryption tax at high frequencies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DME moves security from the "Protocol" layer to the "Physical" layer. It assumes process-level isolation can be bypassed and relies on the CPU's enclave primitives.
* **Observability:** Memory-broker latency and jitter-injection metrics are surfaced in the Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-04:** Initial Document Creation.
