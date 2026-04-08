# Design Doc: Recursive Scratchpad Isolation (RSI) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Parallel agent teams (e.g., Claude Code Agent Teams) often use shared scratchpads to coordinate state. However, today's "Teammate Mailbox" and "Scratchpad" models frequently suffer from "Reasoning Leakage," where intermediate "thought-blobs" from one teammate distract or mislead others before they are verified or committed.

The RSI Manager provides a multi-tenant, sharded scratchpad architecture. It ensures that subagents can "think in private" within an isolated, ephemeral scratchpad shard, which is only merged into the shared teammate mesh after passing a mission-root integrity quorum.

## 2. Goals & Non-Goals
* **Goals:**
    * Host ephemeral, mission-bound scratchpad shards for every parallel subagent.
    * Prevent cross-shard visibility until a "Commit-to-Shared" signal is received.
    * Automatically redact mission-root intents from shards before merging (via RAR Engine integration).
    * Enforce hardware-attested monotonic timestamps for all scratchpad writes.
* **Non-Goals:**
    * Providing long-term persistent storage (RSI shards are ephemeral).
    * Encrypting the blackboard (RSI is focused on *isolation during reasoning*).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 3 parallel coding agents without their intermediate "broken" code fragments polluting each other's attention.
* **The Happy Path (Tasks):**
    1. Parent agent spawns 3 teammates in an RSI-enabled mission.
    2. Each teammate is assigned a unique RSI Shard ID.
    3. Teammate A writes a draft fix to its RSI Shard.
    4. Teammate B queries the shared `.scratchpad` but cannot see Teammate A's draft.
    5. Teammate A completes the fix and sends a `commit_shard` request.
    6. RSI Manager triggers the RAR Engine to redact internal reasoning.
    7. The sanitized fix is merged into the shared teammate `.scratchpad`.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent A] -->|Write| B[RSI Shard A]
        C[Subagent B] -->|Read Shared| D[.scratchpad]
        B -->|Uncommitted| E{Isolation Wall}
        E -->|Blocks Read| C
        A -->|Commit Signal| F[RAR Sanitizer]
        F -->|Merged Data| D
    ```
* **APIs / Interfaces:**
    * `WriteShard(shardID string, content string)`
    * `CommitShard(shardID string, targetID string)`
* **Data Storage/State:**
    * Shards are stored in memory-mapped regions with kernel-level isolation (memfd-bound).

## 5. Alternatives Considered
* **Directory-based Isolation**: Rejected because it's difficult to enforce with standard OS permissions in containerized swarms.
* **Branch-based Git Handoffs**: Rejected due to high latency (1s+) for high-frequency teammate coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shard access is bound to hardware-attested subagent tokens.
* **Observability:** Shard commit events and "Isolation Faults" (unauthorized read attempts) are logged.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
