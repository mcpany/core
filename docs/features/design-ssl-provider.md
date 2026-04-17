# Design Doc: Segregated State Lease (SSL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents operate within high-density swarms, they often share state through scratchpads and task lists. However, when a specific mission hardware lease (MBHL) expires, the associated state fragments often remain in memory ("State Ghosting"), potentially leaking sensitive mission data to subsequent tasks or different agents.

The Segregated State Lease (SSL) Provider ensures that all mission-local state is cryptographically bound to a hardware lease, enabling automatic and verifiable isolation or purging upon lease expiration.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically isolate mission-local memory shards (scratchpads, mailboxes).
    * Bind state accessibility to active hardware-locked mission leases.
    * Automate the purging or locking of state fragments upon lease termination.
    * Neutralize "State Ghosting" in shared agent environments.
* **Non-Goals:**
    * Managing long-term archival storage (cold storage).
    * Encrypting data at rest on non-volatile media (focused on active session memory).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Auditor
* **Primary Goal:** Ensure that a "Database Specialist" agent cannot access the scratchpad memory of a "Security Auditor" agent after the auditor's mission has concluded.
* **The Happy Path (Tasks):**
    1. Auditor Agent initiates a mission with a hardware lease from the HLML Provider.
    2. SSL Provider creates a unique, cryptographically segregated memory shard for this lease.
    3. Auditor writes sensitive findings to its local scratchpad.
    4. Mission concludes and the hardware lease expires.
    5. SSL Provider immediately rotates the shard keys, making the memory fragment inaccessible even if the process remains active.
    6. A subsequent specialist agent attempts to read the same memory region and receives an "Access Denied" or "Zeroed Fragment" signal.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Hardware Lease] --> B[SSL Provider]
        B -->|Derive Keys| C[Memory Shard]
        C --> D[Agent Workspace]
        A -->|Expire| B
        B -->|Zero/Rotate| C
    ```
* **APIs / Interfaces:**
    * `ssl.AllocateShard(leaseID) -> ShardID`
    * `ssl.WriteFragment(shardID, key, value) -> Status`
    * `ssl.PurgeLeaseState(leaseID) -> Result`
* **Data Storage/State:**
    * **Ephemeral Shard Map:** In-memory registry of active shards and their associated cryptographic mission bindings.

## 5. Alternatives Considered
* **Namespace Isolation (Docker/K8s):** Rejected as too heavy-weight for granular, sub-second agent tasks within a single container.
* **Standard SQLite Row-Level Security:** Insufficient because it doesn't handle in-memory "Ghost Fragments" that might persist in application-level buffers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All memory access requires a valid, hardware-attested lease token.
* **Observability:** Integrated with the "Lease-Bound State Inspector" UI for visualizing shard isolation.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
