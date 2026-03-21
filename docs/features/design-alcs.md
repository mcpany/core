# Design Doc: Attention-Locked Context Sharding (ALCS)

**Status:** Draft **Created:** [2026-06-15]

## 1. Context and Scope

In deep agentic swarms, mission-critical intent fragments are often evicted from
the context window due to "Attention Entropy" caused by high-volume subagent
outputs. ALCS provides a mechanism to "pin" specific context shards to the LLM's
attention layers using hardware-bound attestation.

## 2. Goals & Non-Goals

- **Goals:**
  - Cryptographically priority-tiering of context shards.
  - Real-time monitoring of attention occupancy.
  - Integration with HAAL (Hardware-Attested Attention Locking).
- **Non-Goals:**
  - Increasing the raw context window size of the LLM.
  - Managing long-term agent memory.

## 3. Critical User Journey (CUJ)

- **User Persona:** Swarm Orchestrator
- **Primary Goal:** Ensure that "Zero-Trust" security directives remain in the
  LLM's focus during recursive tool execution.
- **The Happy Path (Tasks):**
  1. Orchestrator marks a security shard as "Mission Root".
  2. ALCS generates a hardware-attested header for the shard.
  3. The shard is injected into the LLM context with "Tier 0" attention
     priority.
  4. ALCS monitors attention weights to ensure Tier 0 fragments remain above the
     eviction threshold.

## 4. Evolutionary Changelog

- **2026-06-15:** Initial Document Creation.
