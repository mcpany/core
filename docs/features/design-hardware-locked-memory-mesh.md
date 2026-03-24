# Design Doc: Hardware-Locked Memory Mesh
**Status:** Draft
**Created:** 2026-05-16

## 1. Context and Scope
With the introduction of Zero-Copy BSH (Binary State Handoff) for parallel teammate coordination, agent swarms now rely on shared-memory regions for sub-millisecond context exchange. However, the disclosure of "Parallel Inode Racing" (PIR) and memory smearing incidents in production swarms demonstrates that logical memory isolation is insufficient. Malicious subagents can bypass software boundaries to read or modify the state of sibling agents during high-concurrency re-alignments.

MCP Any needs to provide a hardware-enforced isolation layer for these shared memory regions. By binding memory shards to hardware security modules (TPM/Secure Enclave), we can ensure that memory is physically restricted to attested agents, neutralizing the risk of cross-agent state injection or exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-backed (TPM/SEP) memory isolation for BSH shared-memory regions.
    * Provide cryptographic attestation for any agent requesting access to a specific memory shard.
    * Prevent "Parallel Inode Racing" by physically locking memory handles to verified agent IDs.
    * Maintain sub-millisecond latency for context handoffs while adding hardware security.
* **Non-Goals:**
    * Replacing existing network-based transport (this is specifically for local, high-density coordination).
    * Providing general-purpose hardware encryption for all agent data (scoped to BSH shards).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Enable 50+ parallel subagents to coordinate via shared memory without risking cross-agent state leakage.
* **The Happy Path (Tasks):**
    1. The Architect configures a "Hardware-Locked Shard" policy in MCP Any.
    2. A Parent Agent spawns three parallel subagents and requests a shared memory shard for coordination.
    3. MCP Any interacts with the local Secure Enclave to provision a memory region bound to the Parent's hardware-attested ID.
    4. Each subagent must present a valid, Parent-signed "Sub-Attestation Token" to MCP Any to map the memory region.
    5. The Secure Enclave validates the token and physically unlocks the memory segment for the subagent's process ID.
    6. Subagents perform zero-copy state handoffs at kernel speeds.
    7. Any attempt by an unauthorized sibling to access the shard results in a hardware-level trap and immediate session termination.

## 4. Design & Architecture
* **System Flow:**
    [Agent Process] <--> [MCP Any BSH Broker] <--> [Hardware Security Module / Kernel Shard Driver]
    1. Agent requests Shard Access with Attestation.
    2. BSH Broker verifies Attestation against Mission Root.
    3. Broker calls TPM/SEP to generate a session-bound Memory Key.
    4. Kernel Driver uses Memory Key to permit mmap() to the specific Shard for the Agent's PID.
* **APIs / Interfaces:**
    * `ProvisionShard(intent_id, size) -> shard_handle`
    * `AttestShardAccess(shard_handle, agent_token) -> memory_key`
    * `MapShard(memory_key) -> raw_pointer`
* **Data Storage/State:**
    * Shard metadata is stored in a hardware-protected segment of the MCP Any internal state.
    * Actual context data resides in volatile RAM, protected by hardware memory-tagging or enclave-paging.

## 5. Alternatives Considered
* **Logical KV Isolation:** Rejected because software-level checks are vulnerable to race conditions (PIR) and kernel-level escapes.
* **Encrypted IPC (Named Pipes):** Rejected for this specific use case due to the serialization/deserialization overhead which causes "Coordination Stall" in high-density swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Access is denied by default at the hardware level. Only cryptographically attested agents can "see" the memory.
* **Observability:** Hardware traps and access violations are logged as "Critical Security Events" with full lineage metadata.

## 7. Evolutionary Changelog
* **2026-05-16:** Initial Document Creation.
