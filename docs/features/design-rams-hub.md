<!--
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 -->

# Design Doc: Reasoning-Aware Memory Segmentation (RAMS) Hub

**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
As AI agent swarms evolve, they increasingly rely on shared memory segments (the "Blackboard") for low-latency context exchange. However, static isolation is no longer sufficient. The emergence of Shadow Memory Exfiltration (SME) vulnerabilities allows compromised subagents to exfiltrate context from siblings by speculatively reading uninitialized memory. MCP Any needs a "Just-in-Time" memory architecture that binds memory access to a verifiable reasoning trace and ensures temporal isolation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement Temporal Memory Isolation for subagent memory shards.
    * Bind shard lifecycle to a hardware-attested reasoning trace.
    * Provide sub-millisecond context swapping between specialized agents.
    * Neutralize speculative reading of uninitialized memory (SME).
* **Non-Goals:**
    * Replacing the primary long-term vector memory (handled by ContextEngine).
    * Providing cross-host shared memory (initially limited to local node).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely share a 50MB context shard between a "Researcher" and a "Coder" agent without exposing it to a concurrent "Audit" agent or allowing speculative reads.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a RAMS session with a signed mission root.
    2. Researcher agent requests a RAMS Shard, providing a signed Reasoning Trace.
    3. RAMS Hub validates the trace and allocates a cryptographically isolated shard.
    4. Researcher writes context and completes subtask.
    5. Coder agent requests the same shard, providing a lineage-proof linked to the Researcher.
    6. RAMS Hub validates the lineage, zeros out uninitialized regions, and rotates the encryption key for the Coder.
    7. Shard is automatically purged and unmapped when the parent mission completes.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>RAMS Hub: Request Shard (Intent + Trace)
        RAMS Hub->>Trust Provider: Validate Trace
        Trust Provider-->>RAMS Hub: Verified (Attested)
        RAMS Hub->>Memory Controller: Allocate Shard & Zero-Init
        Memory Controller-->>Agent: Shard Handle (Encrypted)
        Agent->>RAMS Hub: Commit/Release
        RAMS Hub->>Memory Controller: Rotate Key / Temporal Purge
    ```
* **APIs / Interfaces:**
    * `POST /v1/rams/shards/allocate`: Requires `reasoning_trace_sig` and `mission_root_id`.
    * `GET /v1/rams/shards/{id}/access`: Rotates access keys and returns a handle for the next agent in the lineage chain.
* **Data Storage/State:**
    * Shards are stored in memory-mapped files (`/dev/shm`) with per-agent encryption keys managed by the RAMS Controller. uninitialized regions are zeroed out before handoff.

## 5. Alternatives Considered
* **Namespace Isolation (cgroups):** Rejected due to the high overhead of container/namespace context switching for sub-millisecond swaps.
* **JSON-RPC State Passing:** Rejected due to "Token Storm" latency in deep agent chains (50MB+ context objects).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shard access is bound to the hardware-attested reasoning lifecycle. Key rotation occurs at every handoff. Zero-initialization prevents SME speculative reads.
* **Observability:** RAMS Hub logs every shard allocation, access, and rotation event to the Local Security Audit Log.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
