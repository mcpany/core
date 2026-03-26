# Design Doc: Context-Aware Shard Isolation (CASI) Middleware
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
In horizontal teams (e.g., Claude Code Agent Teams), multiple agents often share a mailbox or blackboard. "Shard Pollution" occurs when reasoning drift in one agent leaks into the shared context, confusing siblings. CASI enforces semantic boundaries between these context shards.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce semantic isolation between teammate context shards.
    * Detect and block "Reasoning Drift" leakage.
    * Provide "Read-Only\" vs \"Read-Write\" shard permissions.
* **Non-Goals:**
    * Implement the underlying KV store (handled by Blackboard).
    * Manage subagent lifecycle (handled by Reaper).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a "Researcher Agent" from polluting the "Coder Agent's" context with irrelevant web-search noise.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns two agents; MCP Any creates two isolated CASI Shards.
    2. Researcher Agent writes to its Shard.
    3. Coder Agent attempts to read Researcher's Shard; CASI enforces "Read-Only" or "Filtered" access based on the mission policy.
    4. CASI monitors writes for "Semantic Drift" and alerts if pollution is detected.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> CASI Middleware -> Blackboard (Sharded Storage)`
* **APIs / Interfaces:**
    * `mount_shard(shard_id, policy)`
    * `sync_shard(shard_id, delta)`
* **Data Storage/State:**
    Shard boundaries are maintained in memory and backed by the Blackboard's SQLite implementation.

## 5. Alternatives Considered
* **Full Context Isolation:** Rejected because teammates *need* to share some state to coordinate. CASI provides "Smart Sharing."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shard access is governed by the Mission Root intent.
* **Observability:** Shard mount/unmount and "Pollution\" alerts are logged.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
