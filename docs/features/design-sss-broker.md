# Design Doc: Skill-State Sovereignty (SSS) Broker
**Status:** Draft
**Created:** 2026-07-17

## 1. Context and Scope
With the introduction of Stateful Skill Persistence (SSP) in OpenClaw v3.5, AI agent tools are transitioning from stateless interfaces to persistent entities with their own internal registries. Without proper governance, this persistent state can become a vector for mission drift, data leakage, or unauthorized cross-session contamination. The SSS Broker provides a hardware-attested, tool-bound storage architecture that ensures skill state remains sovereign to the user's mission root and strictly isolated from other agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested, cryptographically isolated storage shards for SSP-compliant skills.
    * Enforce mission-root security policies on all tool-managed persistent state.
    * Neutralize "State Hijacking" where one agent attempts to mutate the persistent state of another tool.
    * Facilitate sub-millisecond state resumption for long-running specialist agents.
* **Non-Goals:**
    * Managing the agent's primary reasoning memory (handled by ContextEngine/Blackboard).
    * Providing general-purpose database services for LLMs.

## 3. Critical User Journey (CUJ)
* **User Persona:** Sovereign Swarm Developer
* **Primary Goal:** Deploy a persistent "Database Admin" skill that maintains its own internal connection pool and schema cache without leaking that state to other subagents.
* **The Happy Path (Tasks):**
    1. The developer registers a tool with `stateful: true` in the MCP manifest.
    2. MCP Any's SSS Broker issues a hardware-attested "State Token" bound to that tool's fingerprint.
    3. The skill executes and writes its internal registry data to the assigned SSS shard.
    4. SSS Broker encrypts the data using a mission-root derived key.
    5. On the next session, the skill resumes and requests its state via the SSS-API, which validates the hardware token before decryption.

## 4. Design & Architecture
* **System Flow:**
  [Stateful Tool] <-> [SSS-API] <-> [SSS Broker (Enclave)] <-> [Encrypted Shard Storage]
* **APIs / Interfaces:**
    * `mcpany.sss.v1.StateBroker`
    * `rpc GetToolState(GetToolStateRequest) returns (GetToolStateResponse)`
    * `rpc PutToolState(PutToolStateRequest) returns (PutToolStateResponse)`
* **Data Storage/State:**
    * Sharded SQLite or KV store, with each shard cryptographically pinned to a (ToolID, MissionRootID) tuple.

## 5. Alternatives Considered
* **Blackboard-Managed Tool State**: Rejected because the Blackboard is designed for agent coordination, not the high-frequency internal metadata management of individual skills.
* **Tool-Local Filesystem Storage**: Rejected as it lacks hardware attestation and centralized security policy enforcement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All state is encrypted at rest using keys rooted in the user's hardware TPM.
* **Observability:** SSS events (reads, writes, rotations) are tracked in the mission-root audit log.

## 7. Evolutionary Changelog
* **2026-07-17:** Initial Document Creation.
