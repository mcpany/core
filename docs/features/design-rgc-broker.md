# Design Doc: Reasoning-Aware Garbage Collection (R-GC) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally and grow deeper in reasoning chains, sharded memory (Blackboard/Context Shards) becomes cluttered with "stale" state. specialist agents often "squat" on high-entropy context fragments long after their specific sub-task is completed, leading to context window exhaustion and semantic noise.

MCP Any needs an automated, reasoning-aware mechanism to prune these fragments based on their actual utility to the current mission root, rather than simple TTL or LRU policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Tie context eviction to the agent's internal reasoning confidence scores.
    * Automatically purge shards when their "Semantic Utility" drops below a hardware-attested threshold.
    * Provide a "Semantic Sweeper" service that periodically audits the mesh for state squatting.
* **Non-Goals:**
    * Replacing OS-level memory management.
    * Pruning fragments that are explicitly marked as "GC-Immune" reasoning anchors.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Maintain high performance in a 20+ agent mesh without manual memory tuning.
* **The Happy Path (Tasks):**
    1. Orchestrator defines a global "Semantic Utility" threshold for a mission.
    2. Subagents report reasoning confidence scores alongside state mutations.
    3. The R-GC Broker monitors shard entropy and subagent confidence.
    4. When a shard's utility score (Confidence / Entropy) falls below the threshold, the Broker issues a "Lease Revocation."
    5. Shard is purged or moved to cold storage (long-term episodic graph).

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> (Reasoning Trace + Confidence) -> [Semantic Utility Monitor]
    [Semantic Utility Monitor] -> (Utility Score) -> [R-GC Broker]
    [R-GC Broker] -> (Purge Command) -> [Sharded Memory / Blackboard]
* **APIs / Interfaces:**
    * `POST /v1/gc/policy`: Set global utility thresholds.
    * `GET /v1/gc/status`: View real-time utility scores of active shards.
* **Data Storage/State:**
    Utility scores are stored in a volatile metadata-cache associated with each memory shard.

## 5. Alternatives Considered
* **Time-to-Live (TTL):** Rejected because mission phases have unpredictable durations; TTL leads to "Context Amnesia" if set too short.
* **LRU (Least Recently Used):** Rejected because some "old" fragments (like Mission Root instructions) are the most important.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Purge" command must be hardware-attested to prevent malicious subagents from "silencing" sibling agents by pruning their state.
* **Observability:** Prometheus metrics for "Reclaimed Token Space" and "Utility Decay Rate."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
