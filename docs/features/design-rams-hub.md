# Design Doc: Reasoning-Aware Memory Segmentation (RAMS) Hub
**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
With the emergence of "Recursive Context Splicing" (RCS) and "Memory Smearing" in deep agent swarms, the traditional monolithic "Blackboard" or Shared KV Store has become a liability. OpenClaw's prototype of "Intent-Bound Memory Isolation" (IBMI) and the new pluggable "ContextEngine" demand that MCP Any evolves its state management layer.

The RAMS Hub is a core architectural evolution that transitions the Blackboard from a flat KV store to a segmented, intent-aware memory architecture. It provides the infrastructure to isolate subagent state into cryptographically sealed shards that are bound to specific reasoning missions.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Intent-Sealed Shards" for subagent state isolation.
    * Provide a pluggable adapter for OpenClaw v2026.3.7 ContextEngine lifecycle hooks.
    * Support cryptographic binding of memory regions to mission-root intents.
    * Enable sub-millisecond memory-mapped handoffs between isolated shards (BSH).
* **Non-Goals:**
    * Implementing a general-purpose database for long-term storage (focus is on active reasoning state).
    * Providing automated "Reasoning" or "Decision Making" (focus is on the infrastructure for state isolation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw User)
* **Primary Goal:** Securely delegate a sub-task to a specialized agent without allowing it to "smear" or exfiltrate the parent's sensitive context.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a "Mission Root" intent via MCP Any.
    2. RAMS Hub generates a primary "Intent-Sealed Shard" for the parent.
    3. Parent delegates a sub-task; RAMS Hub automatically spawns a child shard, cryptographically bound to the sub-intent.
    4. Child agent writes to its shard; data is invisible to siblings and read-only for the parent until explicitly merged.
    5. Upon sub-task completion, the ContextEngine lifecycle hook triggers a "Snapshot-and-Merge" into the parent shard, validated by the Semantic Integrity Bridge.

## 4. Design & Architecture
* **System Flow:**
    * **Intent Mesh**: The top-level graph that maps mission-root intents to specific shards.
    * **Shard Manager**: Handles the lifecycle (Mount/Unmount/Seal) of memory-mapped regions.
    * **Cryptographic Layer**: Uses session-bound keys (derived from HEPA/TPM) to seal shard contents.
    * **ContextEngine Bridge**: Translates OpenClaw lifecycle signals (`on_context_create`, `on_context_retrieve`) into Shard Manager actions.
* **APIs / Interfaces:**
    * `POST /v1/rams/shards`: Create a new intent-sealed shard.
    * `GET /v1/rams/shards/{id}/access`: Request a time-bound, intent-scoped access token.
    * `POST /v1/rams/shards/{id}/merge`: Trigger a validated merge into a parent shard.
* **Data Storage/State:**
    * Active shards reside in memory-mapped files (Shadow-FS) for zero-copy transport.
    * Metadata and Intent-Lineage are persisted in an embedded SQLite database (The "Blackboard Meta-Store").

## 5. Alternatives Considered
* **Flat SQLite with Row-Level Security (RLS)**: Rejected due to performance overhead in high-frequency swarms and lack of "Hardware-Sealing" capabilities.
* **Purely Local-Variable State**: Rejected as it prevents state sharing between framework-neutral agents (e.g., OpenClaw to AutoGen).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * **Isolation**: Shards are isolated at the process/memory level.
    * **Attestation**: Every shard mount requires a valid PoI (Proof-of-Intent) and HEPA attestation.
* **Observability:**
    * **RAMS Shard Inspector**: UI for visualizing shard boundaries and intent lineage.
    * **Memory Smearing Alerts**: Real-time alerts when a shard mutation exceeds its signed intent boundary.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
