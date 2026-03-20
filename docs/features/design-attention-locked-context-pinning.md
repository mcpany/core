# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Attention-Locked Context Pinning (ALCP)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As context windows grow, agents are increasingly susceptible to **Context Window Flooding (CWF)**, where malicious subagents or noisy tool outputs inject high-entropy data designed to evict critical instructions or mission-root constraints from the model's active attention. ALCP provides a mechanism to "lock" specific context fragments into high-priority attention tiers using emerging hardware-bound model headers.

## 2. Goals & Non-Goals
* **Goals:**
    * Allow the Mission Root to designate specific fragments (e.g., system prompts, core constraints) as "Attention-Locked."
    * Interface with model-specific attention-tiering APIs (e.g., hardware-bound KV cache pinning).
    * Automatically detect and alert on "Attention Squatting" attempts by subagents.
* **Non-Goals:**
    * Expanding the physical context window size.
    * Managing long-term memory retrieval (handled by ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Sovereignty Mission Root Agent
* **Primary Goal:** Ensure that core security constraints cannot be "flushed" out of the model's attention by a noisy specialized subagent.
* **The Happy Path (Tasks):**
    1. Mission Root initiates a session and marks the "Security Policy" fragment with an `Attention-Lock` token.
    2. A specialized "Log Analyzer" subagent returns 500KB of raw log data.
    3. The Context Shifter prepares the next prompt.
    4. ALCP ensures the `Attention-Lock` fragment is kept in the model's "High-Priority" KV cache, even if other fragments must be evicted or summarized.
    5. The model retains full awareness of the Security Policy despite the log flood.

## 4. Design & Architecture
* **System Flow:**
    `Context Fragment` -> `ALCP Labeler` -> `Attention-Aware Orchestrator` -> `Model API (with Attention Headers)`
* **APIs / Interfaces:**
    * `IAttentionManager`: Service to manage priority tiers for context fragments.
    * `AttentionHeaderProvider`: Translates internal lock status to model-specific API headers.
* **Data Storage/State:**
    * Lock status is stored as metadata on context shards in the `Live Context Shard Manager`.

## 5. Alternatives Considered
* **Recursive Summarization:** Rejected as a primary solution because summarization can lose the semantic precision required for security constraints.
* **Context Budgeting:** Complementary, but doesn't solve the problem of high-priority data being evicted by low-priority data within a full window.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Only agents with "Supervisor" or "Mission Root" lineage can issue Attention-Locks.
* **Observability:** Metrics on `mcpany_attention_locked_fragments_count` and `mcpany_attention_eviction_events_total`.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
