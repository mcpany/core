# Design Doc: Memory-Mapped Intent Isolator (MMII)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
With the introduction of Memory-Mapped Intent Persistence (MMIP) in OpenClaw v3.2.0, agent state transfer has reached sub-millisecond speeds by utilizing shared memory regions. However, this has created a critical "Ghost-State" vulnerability where terminated subagents leave unauthorized intent fragments in the shared buffer. MCP Any needs to provide a secure, hardware-attested memory management layer that ensures absolute state sovereignty between agent sessions.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested, session-bound memory regions for agent state.
    * Ensure atomic purging of memory fragments upon agent termination.
    * Provide cryptographic proof of memory isolation to the mission root.
* **Non-Goals:**
    * Implementing the underlying shared memory primitive (this relies on the OS/Docker).
    * Managing non-agent-related system memory.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Securely share state between 5 specialized subagents using MMIP without risk of "Ghost-State" leakage.
* **The Happy Path (Tasks):**
    1. The Orchestrator requests a hardware-attested memory region from the MMII.
    2. The MMII issues a session-bound memory token cryptographically linked to the Mission Root.
    3. Subagents use the token to mount the isolated memory shard for BSH.
    4. Upon sub-task completion, the MMII receives a termination signal.
    5. The MMII forcefully purges the memory region and invalidates the session token.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Request Region -> MMII (TPM Auth) -> Shard Created -> Subagent Access -> Mission End -> MMII Purge`
* **APIs / Interfaces:**
    * `POST /v1/memory/shard`: Create a session-bound isolated shard.
    * `DELETE /v1/memory/shard/{id}`: Forcefully purge and invalidate a shard.
* **Data Storage/State:**
    * Shard metadata stored in the Shared KV Store (Blackboard).
    * Actual state resides in hardware-bound shared memory buffers.

## 5. Alternatives Considered
* **Stdio-based BSH:** Rejected due to 100ms+ serialization latency in deep swarms.
* **Global Memory Lock:** Rejected as it causes "Cognitive Stall" in parallel Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All memory access requires a hardware-attested token; shards are cryptographically bound to a unique PID and session ID.
* **Observability:** Real-time metrics for memory pressure and "Ghost-Fragment" detection events.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
