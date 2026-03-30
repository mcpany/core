# Design Doc: Enclave-Bound Coordination Locks (EBCL)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As horizontal agent swarms scale to high densities, coordination bottlenecks often occur at the software layer. Even with Conflict-Free Replicated Data Types (CRDTs), rapid concurrent updates to the same state shard can lead to high CPU overhead and potential "Shard-Claim Collision" vulnerabilities.

Enclave-Bound Coordination Locks (EBCL) move the critical synchronization logic into the Hardware Security Module (TPM/Secure Enclave). By using hardware-bound mutexes, MCP Any can ensure that inter-teammate coordination is both race-free and cryptographically non-repudiable at the hardware level.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-enclave (TPM) bound mutexes for coordination shards.
    * Neutralize "Shard-Claim Collision" vulnerabilities in high-density horizontal swarms.
    * Ensure that only hardware-attested identities can acquire coordination locks.
    * Achieve sub-millisecond lock acquisition latency via local bus attestation.
* **Non-Goals:**
    * Replacing software-level CRDTs (EBCL provides the lock; CRDTs manage the data).
    * Providing long-term persistent locks (EBCL locks are mission-session bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** 20 parallel agents claiming tasks from a shared CRDT buffer without software-layer race conditions.
* **The Happy Path (Tasks):**
    1. Agent A identifies a task and requests an EBCL lock for Shard-7.
    2. The EBCL Provider verifies Agent A's hardware-attested identity.
    3. The TPM issues a signed "Lock Acquisition Token" for Shard-7.
    4. Agent B simultaneously requests the same lock; the TPM rejects the request as "Occupied."
    5. Agent A updates the CRDT buffer and releases the lock via a hardware-signed signal.
    6. Agent B is notified of lock availability and proceeds.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] <-> [EBCL Middleware] <-> [TPM/Secure Enclave (Lock Registry)]`
* **APIs / Interfaces:**
    * `AcquireLock(shard_id, identity_token) -> (lock_token, error)`
    * `ReleaseLock(shard_id, lock_token) -> error`
    * `GetLockStatus(shard_id) -> LockInfo`
* **Data Storage/State:**
    * Lock registry resides in hardware-protected volatile memory within the enclave.
    * Non-sensitive lock metadata is mirrored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Software-only Mutex (Sync.Mutex):** Rejected because it doesn't provide the cryptographic non-repudiation required for Zero-Trust inter-agent coordination.
* **Redis-based Distributed Locks:** Rejected due to the 5-10ms network latency tax which is unacceptable for local horizontal swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** EBCL tokens are time-bound and cryptographically linked to the specific mission-root intent.
* **Observability:** Lock acquisition success/fail rates and contention metrics are visualized in the "Shard Lock Visualizer."

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
