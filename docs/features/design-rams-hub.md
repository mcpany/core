<!--
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 -->

# Design Doc: RAMS Hub (Reasoning-Aware Memory Segmentation)
**Status:** Draft | In Review | Approved
**Created:** 2026-05-06

## 1. Context and Scope
MCP Any needs a way to manage shared context across multiple agents in a swarm without exposing the entire memory space to every subagent. The "Shadow Memory Exfiltration" (SME) exploit has shown that naive shared memory is a major security risk. The RAMS Hub solves this by sharding memory and binding its lifecycle to a verifiable "Reasoning Trace."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement temporal memory isolation (ephemeral memory segments).
    * Bind memory access to the active task reasoning trace.
    * Provide hardware-attested memory sharding for high-security environments.
* **Non-Goals:**
    * Long-term vector database storage (out of scope for the gateway).
    * End-to-end encryption of the underlying memory (handled at the storage layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local environment variables or sensitive mission data to unauthorized sub-tasks.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a new RAMS Shard for the "Project Build" mission.
    2. Orchestrator binds the shard to a Reasoning Trace hash.
    3. Subagent A requests access to the shard by presenting a matching Reasoning Trace part.
    4. RAMS Hub validates the trace and grants temporary access.
    5. Shard is automatically pruned once the reasoning trace is finalized (mission complete).

## 4. Design & Architecture
* **System Flow:**
    `[Orchestrator] -> [Reasoning Proof] -> [RAMS Hub (Validator)] -> [Memory Shard]`
* **APIs / Interfaces:**
    * `POST /v1/rams/shard/init` - Create a new isolated memory segment.
    * `POST /v1/rams/shard/grant` - Grant access based on a reasoning trace part.
    * `GET /v1/rams/shard/read` - Read-only access for authorized subagents.
* **Data Storage/State:** State is managed via an in-memory TTL-bound LRU cache, backed by the "Reasoning Trace" as the primary key.

## 5. Alternatives Considered
* **Global Shared Memory**: Rejected due to SME vulnerability and high blast radius for prompt injection.
* **Agent-Specific Databases**: Rejected due to high latency and complexity in synchronizing state for real-time swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses Reasoning-Bound handshakes. No access is granted without a verifiable intent justified by the active mission trace.
* **Observability:** All shard access and pruning events are logged with the associated Reasoning Trace hash for auditability.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
* **2026-05-06 Update: Mitigating Shadow Memory Exfiltration (SME)**: Introducing temporal isolation where memory segments are bound to the hardware-attested Reasoning Trace lifecycle.
